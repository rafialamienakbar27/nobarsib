package domain

import (
	"errors"
	"testing"
)

// State machine §4.5:
//
//	DRAFT ──> PENDING_REVIEW ──> PUBLISHED ──> CONFIRMED ──> FINISHED
//	              │                  │             │
//	              └──> REJECTED      └──> CANCELLED└──> CANCELLED
func TestTransisiYangDiizinkan(t *testing.T) {
	sah := []struct{ from, to string }{
		{EventDraft, EventPendingReview},
		{EventPendingReview, EventPublished},
		{EventPendingReview, EventRejected},
		{EventPublished, EventConfirmed},
		{EventPublished, EventCancelled},
		{EventPublished, EventFinished},
		{EventConfirmed, EventFinished},
		{EventConfirmed, EventCancelled},
	}
	for _, c := range sah {
		if err := ValidateTransition(c.from, c.to); err != nil {
			t.Errorf("%s -> %s seharusnya boleh: %v", c.from, c.to, err)
		}
	}
}

func TestTransisiYangDitolak(t *testing.T) {
	tolak := []struct {
		from, to, alasan string
	}{
		{EventDraft, EventPublished, "draft tidak boleh melompati tinjauan admin"},
		{EventDraft, EventConfirmed, "draft tidak boleh langsung terkonfirmasi"},
		{EventPendingReview, EventConfirmed, "belum ditinjau tapi sudah dikonfirmasi"},
		{EventRejected, EventPublished, "yang sudah ditolak tidak bisa dihidupkan"},
		{EventCancelled, EventPublished, "yang sudah dibatalkan tidak bisa dihidupkan"},
		{EventFinished, EventConfirmed, "laga sudah lewat"},
		{EventPublished, EventPendingReview, "tidak boleh mundur ke antrian"},
		{EventConfirmed, EventPublished, "konfirmasi tidak boleh dicabut diam-diam"},
	}
	for _, c := range tolak {
		err := ValidateTransition(c.from, c.to)
		if err == nil {
			t.Errorf("%s -> %s seharusnya ditolak (%s)", c.from, c.to, c.alasan)
			continue
		}
		if !errors.Is(err, ErrInvalidTransition) {
			t.Errorf("%s -> %s: error bukan ErrInvalidTransition: %v", c.from, c.to, err)
		}
	}
}

func TestTransisiKeStatusYangSama(t *testing.T) {
	if err := ValidateTransition(EventPublished, EventPublished); err == nil {
		t.Error("perpindahan ke status yang sama seharusnya ditolak")
	}
}

func TestTransisiDariStatusTidakDikenal(t *testing.T) {
	if err := ValidateTransition("PUBLISHED", EventConfirmed); err == nil {
		t.Error("status kapital bukan nilai yang disimpan; seharusnya ditolak")
	}
}

// Status akhir tidak boleh punya jalan keluar. Kalau daftar transisi diperluas
// suatu saat, test ini yang pertama gagal.
func TestStatusAkhirTidakPunyaTujuan(t *testing.T) {
	for _, s := range []string{EventRejected, EventCancelled, EventFinished} {
		for _, to := range []string{EventDraft, EventPendingReview, EventPublished, EventConfirmed} {
			if CanTransition(s, to) {
				t.Errorf("%s adalah status akhir tapi bisa pindah ke %s", s, to)
			}
		}
	}
}

func TestValidateEvent(t *testing.T) {
	cases := []struct {
		nama    string
		event   NobarEvent
		wantErr bool
	}{
		{"min_order wajar", NobarEvent{EntryType: EntryMinOrder, EntryAmount: 25000}, false},
		{"gratis dengan nol", NobarEvent{EntryType: EntryFree, EntryAmount: 0}, false},
		// Kartu venue tidak boleh menampilkan "Gratis" dan "Rp 25.000" sekaligus.
		{"gratis tapi berbayar", NobarEvent{EntryType: EntryFree, EntryAmount: 25000}, true},
		{"jumlah negatif", NobarEvent{EntryType: EntryTicket, EntryAmount: -1}, true},
		{"entry_type ngawur", NobarEvent{EntryType: "bayar"}, true},
		{"crowd_level ngawur", NobarEvent{EntryType: EntryFree, CrowdLevel: "sepi"}, true},
		{"crowd_level kosong", NobarEvent{EntryType: EntryFree, CrowdLevel: ""}, false},
		{"crowd_level sah", NobarEvent{EntryType: EntryFree, CrowdLevel: "penuh"}, false},
	}
	for _, c := range cases {
		err := c.event.Validate()
		if (err != nil) != c.wantErr {
			t.Errorf("%s: Validate() = %v, mau error=%v", c.nama, err, c.wantErr)
		}
	}
}
