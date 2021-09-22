package utility

func (u Utils) CopyWithNewScale(scale float64) Utils {

	return Utils{
		hasSecretKey:        u.hasSecretKey,
		bootstrapEnabled:    u.bootstrapEnabled,
		BootstrappingParams: u.BootstrappingParams,
		Params:              u.Params,
		KeyChain:            u.KeyChain,
		Bootstrapper:        u.Bootstrapper,
		Encoder:             u.Encoder,
		Evaluator:           u.Evaluator,
		Encryptor:           u.Encryptor,
		Decryptor:           u.Decryptor,
		Filters:             u.Filters,
		Scale:               scale,
		log:                 u.log,
	}

}

func (u Utils) CopyWithClonedEval() Utils {

	return Utils{
		hasSecretKey:        u.hasSecretKey,
		bootstrapEnabled:    u.bootstrapEnabled,
		BootstrappingParams: u.BootstrappingParams,
		Params:              u.Params,
		KeyChain:            u.KeyChain,
		Bootstrapper:        u.Bootstrapper,
		Encoder:             u.Encoder,
		Evaluator:           u.Evaluator.ShallowCopy(),
		Encryptor:           u.Encryptor,
		Decryptor:           u.Decryptor,
		Filters:             u.Filters,
		Scale:               u.Scale,
		log:                 u.log,
	}

}

func (u Utils) ShallowCopy() Utils {

	return Utils{
		hasSecretKey:        u.hasSecretKey,
		bootstrapEnabled:    u.bootstrapEnabled,
		BootstrappingParams: u.BootstrappingParams,
		Params:              u.Params,
		KeyChain:            u.KeyChain,
		Bootstrapper:        u.Bootstrapper.ShallowCopy(),
		Encoder:             u.Encoder,
		Evaluator:           u.Evaluator.ShallowCopy(),
		Encryptor:           u.Encryptor,
		Decryptor:           u.Decryptor,
		Filters:             u.Filters,
		Scale:               u.Scale,
		log:                 u.log,
	}

}
