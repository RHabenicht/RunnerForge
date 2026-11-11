//go:build windows

package main

import (
	"nsrcgii/srcapp"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNativePDFExport(t *testing.T) {
	ch := srcapp.NewCharacter()
	ch.Name = "José — RunnerForge PDF"
	ch.Notes = "Unicode and complete sheet export"
	dir := filepath.Join(t.TempDir(), "space # unicode é")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "sheet.pdf")
	if err := os.WriteFile(target, []byte("stale file"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := pdfFromHTML(ch, target); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(target)
	if err != nil || len(data) < 1000 || !strings.HasPrefix(string(data), "%PDF-") {
		t.Fatalf("invalid PDF: %v", err)
	}
}
