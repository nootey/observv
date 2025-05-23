package simulate

import (
	"github.com/nootey/observv/internal/bootstrap"
	"go.uber.org/zap"
	"math/rand"
	"time"
)

func StartFakeTraffic(container *bootstrap.Container, logger *zap.Logger) {

	terminals, err := container.TerminalRepo.FindAllTerminals()
	if err != nil {
		logger.Error("Failed to query terminals", zap.Error(err))
		return
	}

	if len(terminals) < 2 {
		logger.Warn("Not enough terminals to simulate traffic. Skipping.", zap.Int("terminal_count", len(terminals)))
		return
	}

	logger.Info("Starting fake traffic simulation", zap.Int("terminal_count", len(terminals)))

	go func() {
		ticker := time.NewTicker(15 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			src := terminals[rand.Intn(len(terminals))]
			dst := terminals[rand.Intn(len(terminals))]
			if src.ID == dst.ID {
				continue
			}

			bytes := int64(rand.Intn(100_000_000)) // 0–100 MB

			err := container.TerminalRepo.UpdateDataSentByID(src.ID, bytes)
			if err != nil {
				logger.Warn("Failed to update terminal data", zap.Uint("terminal_id", src.ID), zap.Error(err))
				continue
			}

			logger.Info("Simulated data transfer",
				zap.Uint("from_terminal_id", src.ID),
				zap.Uint("to_terminal_id", dst.ID),
				zap.Int64("bytes_transferred", bytes),
			)

		}
	}()
}

func StartStatusToggleLoop(container *bootstrap.Container, logger *zap.Logger) {
	logger.Info("Starting fake status change simulation")

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			terminals, err := container.TerminalRepo.FindAllTerminals()
			if err != nil {
				logger.Warn("Failed to refresh terminals for status toggle", zap.Error(err))
				continue
			}

			if len(terminals) == 0 {
				logger.Warn("No terminals available to toggle")
				continue
			}

			// Randomly shuffle terminals
			rand.Shuffle(len(terminals), func(i, j int) {
				terminals[i], terminals[j] = terminals[j], terminals[i]
			})

			// Pick some terminals to toggle
			percent := rand.Intn(41) + 30 // 30–70%
			countToToggle := len(terminals) * percent / 100
			toToggle := terminals[:countToToggle]

			for _, terminal := range toToggle {
				newStatus := "online"
				if terminal.Status == "online" {
					newStatus = "offline"
				}

				err := container.TerminalRepo.UpdateStatusByID(terminal.ID, newStatus)
				if err != nil {
					logger.Warn("Failed to update terminal status", zap.Uint("terminal_id", terminal.ID), zap.Error(err))
					continue
				}

				logger.Debug("Toggled terminal status",
					zap.Uint("terminal_id", terminal.ID),
					zap.String("new_status", newStatus),
				)
			}

			logger.Info("Status toggle tick complete",
				zap.Int("total_terminals", len(terminals)),
				zap.Int("toggled", countToToggle),
			)
		}
	}()
}
