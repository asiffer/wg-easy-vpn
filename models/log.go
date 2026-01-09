package models

import "github.com/rs/zerolog/log"

var logger = log.With().Str("module", "models").Logger()
