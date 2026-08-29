package messenger

import "context"

// Phase151PendingRecoveryOnceAdapter provides one explicit recovery-pass
// boundary without scheduling or installing recovery into the application.
type Phase151PendingRecoveryOnceAdapter struct {
	repository Phase151PendingRecoveryRepository
	config     Phase151PendingRecoveryConfig
}

func NewPhase151PendingRecoveryOnceAdapter(
	repository Phase151PendingRecoveryRepository,
	config Phase151PendingRecoveryConfig,
) *Phase151PendingRecoveryOnceAdapter {
	if phase151PostCreateDependencyNil(repository) || !phase151PendingRecoveryOnceConfigValid(config) {
		return nil
	}
	return &Phase151PendingRecoveryOnceAdapter{repository: repository, config: config}
}

func (a *Phase151PendingRecoveryOnceAdapter) Execute(ctx context.Context) error {
	if a == nil || phase151PostCreateDependencyNil(ctx) ||
		phase151PostCreateDependencyNil(a.repository) || !phase151PendingRecoveryOnceConfigValid(a.config) {
		return ErrPostCreatePipelineInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	return ExportPhase151ExecutePendingPostCreateRecovery(ctx, a.repository, a.config)
}

func (a *Phase151PendingRecoveryOnceAdapter) PendingRecoveryConfig() Phase151PendingRecoveryConfig {
	if a == nil {
		return Phase151PendingRecoveryConfig{}
	}
	return a.config
}

func phase151PendingRecoveryOnceConfigValid(config Phase151PendingRecoveryConfig) bool {
	return config.MaxPages > 0 && config.MaxTargets > 0 &&
		config.PendingScanLimit > 0 && config.MinimumInterval > 0
}
