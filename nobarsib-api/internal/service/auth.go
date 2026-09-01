package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"

	"github.com/nobarsib/nobarsib-api/internal/domain"
)

// Umur token.
//
// Access token dibuat pendek karena tidak bisa dicabut sebelum kedaluwarsa;
// refresh token panjang tapi tersimpan di Redis, sehingga bisa dibatalkan
// seketika saat logout atau kalau ada kecurigaan.
const (
	AccessTokenTTL  = 15 * time.Minute
	RefreshTokenTTL = 7 * 24 * time.Hour
)

// bcryptCost 12 dipilih, bukan default 10. Login admin jarang terjadi, jadi
// tambahan ~200 ms tidak terasa oleh siapa pun kecuali penyerang yang mencoba
// menebak kata sandi.
const bcryptCost = 12

var ErrKredensialSalah = errors.New("email atau kata sandi salah")

type AuthService struct {
	users     domain.UserRepository
	rdb       *redis.Client
	jwtSecret []byte
}

func NewAuthService(users domain.UserRepository, rdb *redis.Client, jwtSecret string) *AuthService {
	return &AuthService{users: users, rdb: rdb, jwtSecret: []byte(jwtSecret)}
}

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int
	User         *domain.User
}

// Login memeriksa kredensial dan menerbitkan sepasang token.
func (s *AuthService) Login(ctx context.Context, email, password string) (*TokenPair, error) {
	u, err := s.users.GetByEmail(ctx, strings.TrimSpace(email))
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			// Waktu tanggapan disamakan dengan kasus kata sandi salah supaya
			// tidak bisa dipakai menebak email mana yang terdaftar.
			bcrypt.CompareHashAndPassword(
				[]byte("$2a$12$aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
				[]byte(password))
			return nil, ErrKredensialSalah
		}
		return nil, err
	}

	if !u.IsActive {
		return nil, ErrKredensialSalah
	}
	if bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)) != nil {
		return nil, ErrKredensialSalah
	}

	if err := s.users.TouchLogin(ctx, u.ID); err != nil {
		return nil, err
	}
	return s.terbitkan(ctx, u)
}

// Refresh menukar refresh token dengan sepasang token baru.
//
// Token lama langsung dihapus (rotasi): kalau sebuah refresh token dipakai dua
// kali, yang kedua pasti gagal — dan itu sinyal token bocor.
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*TokenPair, error) {
	key := refreshKey(refreshToken)

	userID, err := s.rdb.GetDel(ctx, key).Result()
	if err != nil {
		return nil, ErrKredensialSalah
	}
	id, err := uuid.Parse(userID)
	if err != nil {
		return nil, ErrKredensialSalah
	}

	u, err := s.users.GetByID(ctx, id)
	if err != nil || !u.IsActive {
		return nil, ErrKredensialSalah
	}
	return s.terbitkan(ctx, u)
}

// Logout mencabut refresh token. Access token yang masih hidup tetap berlaku
// sampai kedaluwarsa — itulah sebabnya umurnya hanya 15 menit.
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.rdb.Del(ctx, refreshKey(refreshToken)).Err()
}

func (s *AuthService) terbitkan(ctx context.Context, u *domain.User) (*TokenPair, error) {
	sekarang := time.Now()
	claims := Claims{
		Role: u.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   u.ID.String(),
			IssuedAt:  jwt.NewNumericDate(sekarang),
			ExpiresAt: jwt.NewNumericDate(sekarang.Add(AccessTokenTTL)),
			Issuer:    "nobarsib",
		},
	}
	access, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("terbitkan access token: %w", err)
	}

	refresh, err := tokenAcak()
	if err != nil {
		return nil, err
	}
	if err := s.rdb.Set(ctx, refreshKey(refresh), u.ID.String(), RefreshTokenTTL).Err(); err != nil {
		return nil, fmt.Errorf("simpan refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresIn:    int(AccessTokenTTL.Seconds()),
		User:         u,
	}, nil
}

// Verify memeriksa access token dan mengembalikan klaimnya.
func (s *AuthService) Verify(token string) (*Claims, error) {
	var claims Claims
	_, err := jwt.ParseWithClaims(token, &claims, func(t *jwt.Token) (any, error) {
		// Algoritma dipaksa HS256. Tanpa pemeriksaan ini, token ber-alg "none"
		// atau RS256 dengan kunci publik palsu bisa diterima.
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("algoritma tanda tangan tidak diharapkan: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	}, jwt.WithIssuer("nobarsib"), jwt.WithValidMethods([]string{"HS256"}))
	if err != nil {
		return nil, ErrKredensialSalah
	}
	return &claims, nil
}

// HashPassword dipakai saat membuat akun (lihat cmd/adminctl).
func HashPassword(password string) (string, error) {
	if len(password) < 12 {
		return "", fmt.Errorf("%w: kata sandi minimal 12 karakter", domain.ErrInvalidInput)
	}
	b, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hash kata sandi: %w", err)
	}
	return string(b), nil
}

func tokenAcak() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("buat token acak: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func refreshKey(token string) string { return "rt:" + token }
