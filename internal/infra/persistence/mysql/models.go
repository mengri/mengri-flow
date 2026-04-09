package mysql

import (
	accountRepository "mengri-flow/internal/infra/persistence/mysql/account_repository"
	activationTokenRepository "mengri-flow/internal/infra/persistence/mysql/activation_token_repository"
	auditEventRepository "mengri-flow/internal/infra/persistence/mysql/audit_event_repository"
	clusterRepository "mengri-flow/internal/infra/persistence/mysql/cluster_repository"
	credentialRepository "mengri-flow/internal/infra/persistence/mysql/credential_repository"
	environmentRepository "mengri-flow/internal/infra/persistence/mysql/environment_repository"
	identityRepository "mengri-flow/internal/infra/persistence/mysql/identity_repository"
	resourceRepository "mengri-flow/internal/infra/persistence/mysql/resource_repository"
	sessionRepository "mengri-flow/internal/infra/persistence/mysql/session_repository"
)

func init() {
	const event = "AutoMigrateOnDebug"

	accountRepository.Auto(event)
	credentialRepository.Auto(event)
	identityRepository.Auto(event)
	activationTokenRepository.Auto(event)
	auditEventRepository.Auto(event)
	sessionRepository.Auto(event)
	clusterRepository.Auto(event)
	environmentRepository.Auto(event)
	resourceRepository.Auto(event)
}
