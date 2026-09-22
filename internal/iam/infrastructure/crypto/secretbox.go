package crypto

import platformsecurity "erp/internal/platform/security"

type SecretBox = platformsecurity.SecretBox

func NewSecretBox(base64Key string) (*SecretBox, error) {
	return platformsecurity.NewSecretBox(base64Key)
}
