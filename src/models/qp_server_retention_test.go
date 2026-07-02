package models

import "testing"

func TestServerRetentionSetters(t *testing.T) {
	s := &QpServer{}
	if s.StoreRetentionDays.Valid || s.DispatchTypes.Valid {
		t.Fatal("defaults must be NULL (inherit)")
	}
	n := int64(30)
	s.SetStoreRetentionDays(&n)
	if !s.StoreRetentionDays.Valid || s.StoreRetentionDays.Int64 != 30 {
		t.Fatal("set 30 failed")
	}
	s.SetStoreRetentionDays(nil)
	if s.StoreRetentionDays.Valid {
		t.Fatal("set nil must clear to NULL")
	}
	v := "text,image"
	s.SetDispatchTypes(&v)
	if !s.DispatchTypes.Valid || s.DispatchTypes.String != "text,image" {
		t.Fatal("set dispatch types failed")
	}
	s.SetDispatchTypes(nil)
	if s.DispatchTypes.Valid {
		t.Fatal("set nil must clear")
	}
}
