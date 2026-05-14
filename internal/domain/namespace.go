package domain

type Namespace struct {
	ID     string
	Name   string
	Config NamespaceConfig
}

type NamespaceConfig struct {
	MFARequired            bool
	AllowedSocialProviders []string
	PasswordPolicy         PasswordPolicy
	IPAllowlist            []string
	IPDenylist             []string
}

type PasswordPolicy struct {
	MinLength        int
	RequireUppercase bool
	RequireLowercase bool
	RequireNumber    bool
	RequireSpecial   bool
	PasswordHistory  int
}
