package mysql

import (
	accountRepository "mengri-flow/internal/infra/persistence/mysql/account_repository"
	activationTokenRepository "mengri-flow/internal/infra/persistence/mysql/activation_token_repository"
	auditEventRepository "mengri-flow/internal/infra/persistence/mysql/audit_event_repository"
	credentialRepository "mengri-flow/internal/infra/persistence/mysql/credential_repository"
	identityRepository "mengri-flow/internal/infra/persistence/mysql/identity_repository"
	sessionRepository "mengri-flow/internal/infra/persistence/mysql/session_repository"
	userRepository "mengri-flow/internal/infra/persistence/mysql/user_respository"
)

func init() {
	const event = "AutoMigrateOnDebug"

	userRepository.Auto(event)
	accountRepository.Auto(event)
	credentialRepository.Auto(event)
	identityRepository.Auto(event)
	activationTokenRepository.Auto(event)
	auditEventRepository.Auto(event)
	sessionRepository.Auto(event)
}
