-- PERBAIKAN #1 (rencana-pengerjaan §Fase 1)
--
-- Blueprint §7.2 mendefinisikan app_user di baris 608, SETELAH tabel venue
-- (baris 475, kolom owner_user_id REFERENCES app_user(id)) dan nobar_event
-- (baris 549, kolom created_by). DDL itu gagal dijalankan apa adanya karena
-- foreign key menunjuk tabel yang belum ada.
--
-- Karena itu app_user dibuat paling awal, sebelum semua tabel yang mereferensinya.

CREATE TABLE app_user (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    email           VARCHAR(150) UNIQUE,
    phone           VARCHAR(30) UNIQUE,
    password_hash   TEXT,
    full_name       VARCHAR(120),
    role            VARCHAR(20) NOT NULL DEFAULT 'venue_owner',
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    last_login_at   TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT app_user_role_valid CHECK (role IN ('admin', 'venue_owner')),
    -- Minimal satu kanal identitas harus ada. Login venue lewat magic link WA
    -- (§15.2) memakai phone, admin memakai email.
    CONSTRAINT app_user_identity_present CHECK (email IS NOT NULL OR phone IS NOT NULL)
);
