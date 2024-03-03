package Service

import (
	_ "bytes"
	"crypto/rand"
	"encoding/base32"
	"image/color"
	_ "image/png"

	"github.com/pquerna/otp/totp"
	qrcode "github.com/skip2/go-qrcode"
)

// TwoFactorAuthenticationService representa el servicio para la autenticación de dos factores.
type TwoFactorAuthenticationService struct{}

// NewTwoFactorAuthenticationService crea una nueva instancia del servicio de autenticación de dos factores.
func NewTwoFactorAuthenticationService() *TwoFactorAuthenticationService {
	return &TwoFactorAuthenticationService{}
}

// GenerateSecretKey genera una nueva clave secreta para la autenticación de dos factores.
func (s *TwoFactorAuthenticationService) GenerateSecretKey() string {
	randomBytes := make([]byte, 10) // 10 bytes generan una clave de longitud adecuada.
	_, err := rand.Read(randomBytes)
	if err != nil {
		// Manejar el error según sea necesario.
		return ""
	}
	return base32.StdEncoding.EncodeToString(randomBytes)
}

// VerifyCode verifica el código proporcionado por el usuario contra la clave secreta.
func (s *TwoFactorAuthenticationService) VerifyCode(userCode string, secretKey string) bool {
	return totp.Validate(userCode, secretKey)
}

// GenerateTotpUrl genera el URL TOTP a partir de la clave secreta, el emisor y el nombre de la cuenta.
func (s *TwoFactorAuthenticationService) GenerateTotpUrl(secretKey string, issuer string, account string) string {
	opts := totp.GenerateOpts{
		Issuer:      issuer,
		AccountName: account,
		Secret:      []byte(secretKey),
	}
	key, err := totp.Generate(opts)
	if err != nil {
		return ""
	}
	return key.URL()
}

func (s *TwoFactorAuthenticationService) GenerateQrCode(totpUrl string) ([]byte, error) {
	qrCode, err := qrcode.New(totpUrl, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	qrCode.ForegroundColor = color.Black
	qrCode.BackgroundColor = color.White

	pngBytes, err := qrCode.PNG(200)
	if err != nil {
		return nil, err
	}

	return pngBytes, nil
}
