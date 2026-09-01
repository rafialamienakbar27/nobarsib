// Command adminctl mengelola akun admin dari baris perintah.
//
// Ada karena akun admin pertama tidak bisa dibuat lewat panel admin — panelnya
// sendiri butuh login. Sengaja tidak dibuatkan endpoint "daftar admin":
// endpoint semacam itu hanya berguna sekali dan berbahaya selamanya.
//
//	make admin-create EMAIL=kamu@contoh.id NAME="Nama Kamu"
package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"golang.org/x/term"

	"github.com/nobarsib/nobarsib-api/internal/config"
	"github.com/nobarsib/nobarsib-api/internal/domain"
	"github.com/nobarsib/nobarsib-api/internal/repository"
	"github.com/nobarsib/nobarsib-api/internal/service"
)

func main() {
	_ = godotenv.Load()

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	if err := run(os.Args[1], os.Args[2:]); err != nil {
		fmt.Fprintln(os.Stderr, "gagal:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, `adminctl — kelola akun admin NOBARSIB

Perintah:
  create <email> [nama lengkap]   buat akun admin baru
  passwd <email>                  ganti kata sandi akun

Kata sandi selalu ditanyakan lewat prompt, tidak lewat argumen, supaya tidak
tersimpan di riwayat shell.`)
}

func run(perintah string, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := repository.NewPostgres(ctx, cfg.DatabaseURL, 2)
	if err != nil {
		return err
	}
	defer db.Close()

	users := repository.NewUserRepo(db)

	switch perintah {
	case "create":
		if len(args) < 1 {
			return errors.New("pakai: adminctl create <email> [nama lengkap]")
		}
		return buat(ctx, users, args[0], strings.Join(args[1:], " "))
	case "passwd":
		return errors.New("belum tersedia; untuk sekarang buat akun baru lalu nonaktifkan yang lama")
	default:
		usage()
		return fmt.Errorf("perintah tidak dikenal: %s", perintah)
	}
}

func buat(ctx context.Context, users domain.UserRepository, email, nama string) error {
	if !strings.Contains(email, "@") {
		return fmt.Errorf("%q tidak terlihat seperti alamat email", email)
	}

	sandi, err := mintaSandi()
	if err != nil {
		return err
	}
	hash, err := service.HashPassword(sandi)
	if err != nil {
		return err
	}

	u := &domain.User{
		Email:        strings.ToLower(strings.TrimSpace(email)),
		FullName:     nama,
		PasswordHash: hash,
		Role:         domain.RoleAdmin,
		IsActive:     true,
	}
	if err := users.Create(ctx, u); err != nil {
		return err
	}

	fmt.Printf("akun admin dibuat: %s (%s)\n", u.Email, u.ID)
	return nil
}

// mintaSandi membaca kata sandi tanpa menampilkannya, lalu meminta
// pengulangan — salah ketik pada kata sandi yang tak terlihat baru ketahuan
// saat gagal login.
func mintaSandi() (string, error) {
	if !term.IsTerminal(int(syscall.Stdin)) {
		// Bukan terminal (misalnya dijalankan dari skrip): baca satu baris.
		s := bufio.NewScanner(os.Stdin)
		if !s.Scan() {
			return "", errors.New("kata sandi tidak terbaca dari stdin")
		}
		return strings.TrimSpace(s.Text()), nil
	}

	fmt.Print("Kata sandi (minimal 12 karakter): ")
	pertama, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("baca kata sandi: %w", err)
	}

	fmt.Print("Ulangi kata sandi: ")
	kedua, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", fmt.Errorf("baca kata sandi: %w", err)
	}

	if string(pertama) != string(kedua) {
		return "", errors.New("kata sandi tidak sama")
	}
	return string(pertama), nil
}
