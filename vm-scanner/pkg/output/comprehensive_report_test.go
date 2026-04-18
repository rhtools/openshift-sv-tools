//go:build integration

package output

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"testing"

	"vm-scanner/pkg/client"

	"github.com/xuri/excelize/v2"
)

var kubeconfig = flag.String("kubeconfig", "", "path to kubeconfig for integration tests")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

var reportCache struct {
	report *ComprehensiveReport
	err    error
	once   sync.Once
}

func getCachedReport(t *testing.T) *ComprehensiveReport {
	t.Helper()
	if *kubeconfig == "" {
		t.Skip("skipping: -kubeconfig not provided")
	}
	reportCache.once.Do(func() {
		c, err := client.NewClusterClient(client.AuthOptions{Kubeconfig: *kubeconfig})
		if err != nil {
			reportCache.err = fmt.Errorf("failed to create client: %w", err)
			return
		}
		reportCache.report, reportCache.err = GenerateComprehensiveReport(context.Background(), c)
	})
	if reportCache.err != nil {
		t.Fatalf("failed to generate report: %v", reportCache.err)
	}
	return reportCache.report
}

func TestComprehensiveReport_GeneratesSuccessfully(t *testing.T) {
	report := getCachedReport(t)

	if report == nil {
		t.Fatal("report is nil")
	}
	if report.GeneratedAt == "" {
		t.Error("GeneratedAt is empty")
	}
	if report.GeneratedBy != "vm-scanner" {
		t.Errorf("GeneratedBy = %q, want 'vm-scanner'", report.GeneratedBy)
	}
	if len(report.VMs) == 0 {
		t.Error("report.VMs has length 0, expected > 0")
	}
	if len(report.Nodes) == 0 {
		t.Error("report.Nodes has length 0, expected > 0")
	}
	if len(report.Storage) == 0 {
		t.Error("report.Storage has length 0, expected > 0")
	}
	if report.Cluster.TotalVMs != len(report.VMs) {
		t.Errorf("report.Cluster.TotalVMs = %d, want %d", report.Cluster.TotalVMs, len(report.VMs))
	}
	if report.Cluster.RunningVMs+report.Cluster.StoppedVMs != report.Cluster.TotalVMs {
		t.Errorf("RunningVMs + StoppedVMs = %d, want %d",
			report.Cluster.RunningVMs+report.Cluster.StoppedVMs, report.Cluster.TotalVMs)
	}
}

func TestComprehensiveReport_XLSXOutput(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "report.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedSheets := []string{"Summary", "Node Hardware", "Virtual Machines", "Storage Classes",
		"VM Disks", "Network Interfaces", "Capacity Planning", "VM Assessment",
		"PVC Inventory", "NAD Inventory", "DataVolumes",
		"Migration Readiness", "Storage Analysis", "Operator Status"}
	sheets := xlsx.GetSheetList()
	if len(sheets) != len(expectedSheets) {
		t.Errorf("got %d sheets, want %d", len(sheets), len(expectedSheets))
	}
	for _, name := range expectedSheets {
		found := false
		for _, s := range sheets {
			if s == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("sheet %q not found in %v", name, sheets)
		}
	}
	for _, name := range sheets {
		rows, err := xlsx.GetRows(name)
		if err != nil {
			t.Errorf("GetRows(%q) failed: %v", name, err)
			continue
		}
		if len(rows) == 0 {
			t.Errorf("sheet %q has no rows", name)
		}
	}
}

func TestComprehensiveReport_VMDisksSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "vm-disks.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"VM Name", "VM Namespace", "Volume Name", "Volume Type",
		"PVC Size (GiB)", "Storage Class", "Guest Mount Point", "Guest FS Type",
		"Guest Total (GiB)", "Guest Used (GiB)", "Guest Free (GiB)", "Guest Usage (%)"}
	rows, err := xlsx.GetRows("VM Disks")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("VM Disks sheet has only %d rows, expected header + data", len(rows))
	}
	for i, h := range expectedHeaders {
		if rows[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, rows[0][i], h)
		}
	}
	for i, row := range rows[1:] {
		if row[0] == "" || row[1] == "" {
			t.Errorf("row %d: VM Name or VM Namespace is empty", i+1)
		}
	}

	vmsWithDisks := 0
	for _, vm := range report.VMs {
		if len(vm.Disks) > 0 {
			vmsWithDisks++
		}
	}
	if vmsWithDisks > 0 && len(rows)-1 < vmsWithDisks {
		t.Errorf("disk rows = %d, want >= %d (VMs with disks)", len(rows)-1, vmsWithDisks)
	}
}

func TestComprehensiveReport_NetworkInterfacesSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "net-if.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"VM Name", "VM Namespace", "Interface Name", "MAC Address",
		"IP Addresses", "Type", "Model", "Network Name", "NAD Name"}
	rows, err := xlsx.GetRows("Network Interfaces")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 1 {
		t.Fatal("Network Interfaces sheet has no rows")
	}
	for i, h := range expectedHeaders {
		if rows[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, rows[0][i], h)
		}
	}
	for i, row := range rows[1:] {
		if row[0] == "" || row[1] == "" {
			t.Errorf("row %d: VM Name or VM Namespace is empty", i+1)
		}
	}

	// Phase 2: if any row has Type non-empty, enrichment is working.
	// Empty Type is valid when VMs don't have explicit interface type set in spec.
	hasType := false
	for _, row := range rows[1:] {
		if len(row) > 5 && row[5] != "" {
			hasType = true
			break
		}
	}
	// Log rather than fail -- Type may be empty depending on cluster VM spec configuration
	if hasType {
		t.Log("Network interface Type enrichment confirmed")
	}
}

func TestComprehensiveReport_CapacityPlanningSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "capacity.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"Node Name", "CPU Cores", "CPU Allocated to VMs", "CPU Overcommit Ratio",
		"Memory Capacity (GiB)", "Memory Allocated to VMs (GiB)", "Memory Overcommit Ratio",
		"VM Count", "Filesystem Used (%)", "Memory Used (%)"}
	rows, err := xlsx.GetRows("Capacity Planning")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 1 {
		t.Fatal("Capacity Planning sheet has no rows")
	}
	for i, h := range expectedHeaders {
		if rows[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, rows[0][i], h)
		}
	}
	if len(rows) != len(report.Nodes)+1 {
		t.Errorf("data rows = %d, want %d (one per node)", len(rows)-1, len(report.Nodes))
	}
}

func TestComprehensiveReport_VMAssessmentSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "vm-assessment.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"VM Name", "Namespace", "Power State", "Guest Agent",
		"Memory Configured (MiB)", "Memory Used (MiB)", "Memory Utilization (%)", "Memory Waste Flag",
		"vCPUs", "Disk Count", "Total Disk Allocated (GiB)", "Total Disk Used (GiB)",
		"Storage Utilization (%)", "OS Detected", "Run Strategy"}
	rows, err := xlsx.GetRows("VM Assessment")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) != len(report.VMs)+1 {
		t.Errorf("rows = %d, want %d (one per VM plus header)", len(rows), len(report.VMs)+1)
	}
	for i, h := range expectedHeaders {
		if rows[0][i] != h {
			t.Errorf("header[%d] = %q, want %q", i, rows[0][i], h)
		}
	}
	for i, row := range rows[1:] {
		if row[3] != "Yes" && row[3] != "No" {
			t.Errorf("row %d: Guest Agent = %q, want Yes or No", i+1, row[3])
		}
		if row[7] != "" && row[7] != "OVERSIZED" {
			t.Errorf("row %d: Memory Waste Flag = %q, want '' or 'OVERSIZED'", i+1, row[7])
		}
	}

	// Phase 2: Run Strategy column must have at least one non-empty value
	hasRunStrategy := false
	for _, row := range rows[1:] {
		if row[14] != "" {
			hasRunStrategy = true
			break
		}
	}
	if !hasRunStrategy {
		t.Error("Run Strategy column is empty for all VMs")
	}

	// At least one VM should have non-zero vCPUs (catches zero-CPU regression)
	hasNonZeroVCPUs := false
	for _, row := range rows[1:] {
		if len(row) > 8 && row[8] != "" && row[8] != "0" {
			hasNonZeroVCPUs = true
			break
		}
	}
	if !hasNonZeroVCPUs {
		t.Error("vCPUs column is zero or empty for all VMs in VM Assessment sheet")
	}
}

func TestComprehensiveReport_CSVOutput(t *testing.T) {
	report := getCached(t)
	tmpDir := t.TempDir()
	csvOut := filepath.Join(tmpDir, "report.csv")
	formatter := NewFormatter("multi-csv", csvOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	expectedFiles := []string{"_summary.csv", "_node-hardware.csv", "_vms.csv", "_storage.csv",
		"_vm-disks.csv", "_network-interfaces.csv", "_capacity-planning.csv", "_vm-assessment.csv",
		"_pvc-inventory.csv", "_nad-inventory.csv", "_datavolumes.csv",
		"_migration-readiness.csv", "_storage-analysis.csv", "_operator-status.csv"}
	for _, suffix := range expectedFiles {
		fname := "report" + suffix
		fpath := filepath.Join(tmpDir, fname)
		if _, err := os.Stat(fpath); os.IsNotExist(err) {
			t.Errorf("expected file %s not found at %s", fname, fpath)
			continue
		}
		f, err := os.Open(fpath)
		if err != nil {
			t.Errorf("failed to open %s: %v", fname, err)
			continue
		}
		reader := csv.NewReader(f)
		records, err := reader.ReadAll()
		f.Close()
		if err != nil {
			t.Errorf("failed to read %s: %v", fname, err)
			continue
		}
		if len(records) < 2 {
			t.Errorf("%s has only %d lines, want header + data", fname, len(records))
		}
	}
}

func getCached(t *testing.T) *ComprehensiveReport {
	return getCachedReport(t)
}

func TestComprehensiveReport_VMTabPhase2Columns(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "vms.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"Run Strategy", "Instance Type", "Preference", "Labels"}
	rows, err := xlsx.GetRows("Virtual Machines")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 2 {
		t.Fatal("Virtual Machines sheet has no data rows")
	}
	headers := rows[0]

	// Find indices for Phase 2 columns
	colIdx := make(map[string]int)
	for i, h := range headers {
		colIdx[h] = i
	}
	for _, h := range expectedHeaders {
		if _, ok := colIdx[h]; !ok {
			t.Errorf("header %q not found in VM tab", h)
		}
	}

	// Assert at least one non-empty value per Phase 2 column
	for _, h := range expectedHeaders {
		idx, ok := colIdx[h]
		if !ok {
			continue
		}
		hasValue := false
		for _, row := range rows[1:] {
			if idx < len(row) && row[idx] != "" {
				hasValue = true
				break
			}
		}
		if !hasValue {
			t.Errorf("column %q is empty for all VMs", h)
		}
	}
}

func TestComprehensiveReport_InstanceTypeResolution(t *testing.T) {
	report := getCached(t)

	var found bool
	for _, vmDetail := range report.VMs {
		if vmDetail.InstanceType == "" {
			continue
		}
		found = true
		if vmDetail.CPUInfo.VCPUs == 0 {
			t.Errorf("VM %s/%s has InstanceType=%q but VCPUs=0", vmDetail.Namespace, vmDetail.Name, vmDetail.InstanceType)
		}
		if vmDetail.MemoryInfo.MemoryConfiguredMiB == 0 {
			t.Errorf("VM %s/%s has InstanceType=%q but MemoryConfiguredMiB=0", vmDetail.Namespace, vmDetail.Name, vmDetail.InstanceType)
		}
	}
	if !found {
		t.Log("no instancetype-based VMs in cluster, skipping content assertions")
	}
}

func TestComprehensiveReport_PVCInventorySheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "pvc.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"PVC Name", "Namespace", "Status", "Capacity (GiB)", "Access Modes",
		"Storage Class", "Volume Mode", "Bound PV", "Owning VM", "Owning VM Namespace", "Created"}
	rows, err := xlsx.GetRows("PVC Inventory")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("PVC Inventory sheet has only %d rows, expected header + data", len(rows))
	}
	for i, h := range expectedHeaders {
		if i >= len(rows[0]) || rows[0][i] != h {
			actual := ""
			if i < len(rows[0]) {
				actual = rows[0][i]
			}
			t.Errorf("header[%d] = %q, want %q", i, actual, h)
		}
	}

	hasBound := false
	hasStorageClass := false
	validStatuses := map[string]bool{"Bound": true, "Pending": true, "Lost": true}
	for _, row := range rows[1:] {
		if len(row) > 2 && row[2] == "Bound" {
			hasBound = true
		}
		if len(row) > 5 && row[5] != "" {
			hasStorageClass = true
		}
		if len(row) > 2 && row[2] != "" && !validStatuses[row[2]] {
			t.Errorf("invalid PVC status: %q", row[2])
		}
	}
	if !hasBound {
		t.Error("no PVC with Status=Bound found")
	}
	if !hasStorageClass {
		t.Error("no PVC with non-empty Storage Class found")
	}
}

func TestComprehensiveReport_NADInventorySheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "nad.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"Name", "Namespace", "Type", "VLAN", "Resource Name", "Created"}
	rows, err := xlsx.GetRows("NAD Inventory")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 1 {
		t.Fatal("NAD Inventory sheet has no rows")
	}
	for i, h := range expectedHeaders {
		if i >= len(rows[0]) || rows[0][i] != h {
			actual := ""
			if i < len(rows[0]) {
				actual = rows[0][i]
			}
			t.Errorf("header[%d] = %q, want %q", i, actual, h)
		}
	}

	if len(rows) > 1 {
		hasType := false
		for _, row := range rows[1:] {
			if len(row) > 2 && row[2] != "" {
				hasType = true
				break
			}
		}
		if !hasType {
			t.Log("NAD Inventory has data rows but no Type values populated")
		}
	} else {
		t.Log("no NADs in cluster, only header row present")
	}
}

func TestComprehensiveReport_DataVolumeSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "dv.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	expectedHeaders := []string{"Name", "Namespace", "Phase", "Progress", "Source Type",
		"Storage Size (GiB)", "Storage Class", "Owning VM", "Created"}
	rows, err := xlsx.GetRows("DataVolumes")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 1 {
		t.Fatal("DataVolumes sheet has no rows")
	}
	for i, h := range expectedHeaders {
		if i >= len(rows[0]) || rows[0][i] != h {
			actual := ""
			if i < len(rows[0]) {
				actual = rows[0][i]
			}
			t.Errorf("header[%d] = %q, want %q", i, actual, h)
		}
	}

	if len(rows) > 1 {
		hasSucceeded := false
		for _, row := range rows[1:] {
			if len(row) > 2 && row[2] == "Succeeded" {
				hasSucceeded = true
				break
			}
		}
		if !hasSucceeded {
			t.Log("no DataVolume with Phase=Succeeded found (may be expected)")
		}
	} else {
		t.Log("no DataVolumes in cluster, only header row present")
	}
}

func TestComprehensiveReport_SummaryPhase3(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "summary.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}

	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()

	rows, err := xlsx.GetRows("Summary")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}

	foundPVCCount := false
	foundPVCStorage := false
	for _, row := range rows {
		if len(row) >= 2 {
			if row[0] == "PVC Count" && row[1] != "0" {
				foundPVCCount = true
			}
			if row[0] == "Total PVC Storage (GiB)" && row[1] != "0.0" {
				foundPVCStorage = true
			}
		}
	}
	if !foundPVCCount {
		t.Error("Summary missing 'PVC Count' row with non-zero value")
	}
	if !foundPVCStorage {
		t.Error("Summary missing 'Total PVC Storage (GiB)' row with non-zero value")
	}
}

func TestComprehensiveReport_MigrationReadinessSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "migration.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()
	expectedHeaders := []string{"VM Name", "Namespace", "Power State", "Live Migratable",
		"Run Strategy", "Eviction Strategy", "Host Devices", "Node Affinity",
		"PVC Access Mode Issue", "Dedicated CPU", "Guest Agent", "Blockers", "Readiness Score"}
	rows, err := xlsx.GetRows("Migration Readiness")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("Migration Readiness sheet has only %d rows, expected header + data", len(rows))
	}
	for i, h := range expectedHeaders {
		if i >= len(rows[0]) || rows[0][i] != h {
			actual := ""
			if i < len(rows[0]) {
				actual = rows[0][i]
			}
			t.Errorf("header[%d] = %q, want %q", i, actual, h)
		}
	}
	if len(rows) != len(report.VMs)+1 {
		t.Errorf("rows = %d, want %d (one per VM plus header)", len(rows), len(report.VMs)+1)
	}
	scorePattern := regexp.MustCompile(`^\d+/\d+$`)
	validLM := map[string]bool{"Yes": true, "No": true, "Unknown": true}
	for i, row := range rows[1:] {
		if len(row) > 12 && !scorePattern.MatchString(row[12]) {
			t.Errorf("row %d: Readiness Score %q does not match X/Y pattern", i+1, row[12])
		}
		if len(row) > 3 && !validLM[row[3]] {
			t.Errorf("row %d: Live Migratable = %q, want Yes/No/Unknown", i+1, row[3])
		}
	}
}

func TestComprehensiveReport_StorageAnalysisSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "storage-analysis.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()
	expectedHeaders := []string{"PVC Name", "Namespace", "Storage Class", "Capacity (GiB)",
		"Access Modes", "Volume Mode", "Status", "Owning VM", "VM Power State",
		"Guest Used (GiB)", "Utilization (%)", "Flag"}
	rows, err := xlsx.GetRows("Storage Analysis")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 2 {
		t.Fatalf("Storage Analysis sheet has only %d rows, expected header + data", len(rows))
	}
	for i, h := range expectedHeaders {
		if i >= len(rows[0]) || rows[0][i] != h {
			actual := ""
			if i < len(rows[0]) {
				actual = rows[0][i]
			}
			t.Errorf("header[%d] = %q, want %q", i, actual, h)
		}
	}
	validFlags := map[string]bool{"": true, "Orphaned": true, "Overprovisioned": true, "Low Utilization": true}
	hasOwningVM := false
	for _, row := range rows[1:] {
		if len(row) > 11 && !validFlags[row[11]] {
			t.Errorf("invalid Flag value: %q", row[11])
		}
		if len(row) > 7 && row[7] != "" {
			hasOwningVM = true
		}
	}
	if !hasOwningVM {
		t.Log("no PVC with non-empty Owning VM found")
	}
}

func TestComprehensiveReport_OperatorStatusSheet(t *testing.T) {
	report := getCached(t)
	xlsxOut := filepath.Join(t.TempDir(), "operators.xlsx")
	formatter := NewFormatter("xlsx", xlsxOut)
	if err := formatter.Format(report); err != nil {
		t.Fatalf("Format failed: %v", err)
	}
	xlsx, err := excelize.OpenFile(xlsxOut)
	if err != nil {
		t.Fatalf("OpenFile failed: %v", err)
	}
	defer xlsx.Close()
	expectedHeaders := []string{"Operator Name", "Source", "Namespace", "Version", "Status", "Health", "Created"}
	rows, err := xlsx.GetRows("Operator Status")
	if err != nil {
		t.Fatalf("GetRows failed: %v", err)
	}
	if len(rows) < 1 {
		t.Fatal("Operator Status sheet has no rows")
	}
	for i, h := range expectedHeaders {
		if i >= len(rows[0]) || rows[0][i] != h {
			actual := ""
			if i < len(rows[0]) {
				actual = rows[0][i]
			}
			t.Errorf("header[%d] = %q, want %q", i, actual, h)
		}
	}
	if len(rows) > 1 {
		hasHealthy := false
		for _, row := range rows[1:] {
			if len(row) > 5 && row[5] == "Healthy" {
				hasHealthy = true
				break
			}
		}
		if !hasHealthy {
			t.Log("no operator with Health=Healthy found (may be expected depending on cluster)")
		}
	} else {
		t.Log("no operators in cluster, only header row present")
	}
}
