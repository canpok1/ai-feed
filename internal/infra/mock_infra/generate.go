package mock_infra

//go:generate mockgen -source=../config.go -destination=mock_config.go -package=mock_infra ConfigRepository
