package pagination

import (
	"net/http/httptest"
	"testing"
)

func TestParseFromQuery_Defaults(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	p := ParseFromQuery(req)
	if p.Page != 1 {
		t.Errorf("expected page=1, got %d", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("expected per_page=20, got %d", p.PerPage)
	}
}

func TestParseFromQuery_CustomValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/?page=3&per_page=50", nil)
	p := ParseFromQuery(req)
	if p.Page != 3 {
		t.Errorf("expected page=3, got %d", p.Page)
	}
	if p.PerPage != 50 {
		t.Errorf("expected per_page=50, got %d", p.PerPage)
	}
}

func TestParseFromQuery_ClampsMax(t *testing.T) {
	req := httptest.NewRequest("GET", "/?per_page=500", nil)
	p := ParseFromQuery(req)
	if p.PerPage != 100 {
		t.Errorf("expected per_page clamped to 100, got %d", p.PerPage)
	}
}

func TestParseFromQuery_InvalidValues(t *testing.T) {
	req := httptest.NewRequest("GET", "/?page=abc&per_page=-5", nil)
	p := ParseFromQuery(req)
	if p.Page != 1 {
		t.Errorf("expected page=1 for invalid, got %d", p.Page)
	}
	if p.PerPage != 20 {
		t.Errorf("expected per_page=20 for negative, got %d", p.PerPage)
	}
}

func TestOffset(t *testing.T) {
	p := Params{Page: 3, PerPage: 20}
	if p.Offset() != 40 {
		t.Errorf("expected offset=40, got %d", p.Offset())
	}
}

func TestNewResponse(t *testing.T) {
	items := []string{"a", "b", "c"}
	resp := NewResponse(items, 50, Params{Page: 2, PerPage: 20})
	if resp.Total != 50 {
		t.Errorf("expected total=50, got %d", resp.Total)
	}
	if resp.TotalPages != 3 {
		t.Errorf("expected total_pages=3, got %d", resp.TotalPages)
	}
	if len(resp.Items) != 3 {
		t.Errorf("expected 3 items, got %d", len(resp.Items))
	}
}

func TestNewResponse_NilItems(t *testing.T) {
	resp := NewResponse[string](nil, 0, Params{Page: 1, PerPage: 20})
	if resp.Items == nil {
		t.Error("expected empty slice, got nil")
	}
}
