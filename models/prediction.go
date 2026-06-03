package models

import (
	"fmt"
	"time"

	"github.com/janimationd/JacuzziBot/utils"
)

type PredictionState int

const (
	BettingOpen PredictionState = iota
	BettingLocked
	PredictionResolved
	PredictionCancelled
)

type PredictionBet struct {
	OutcomeId rune
	Wager     float64
}

type Prediction struct {
	// The unique ID of the prediction
	Id JacuzziId
	// Who created the prediction? (user ID)
	Creator string
	// The question/text of the prediction.
	Question string
	// When the prediction was created.
	CreationTime time.Time
	// When bets can no longer be placed.
	BettingCloseTime time.Time
	// The ID of the event that has been scheduled to close betting.
	BettingCloseEventId JacuzziId
	// When the prediction will automatically be cancelled (no longer able to be resolved, bets are refunded).
	ExpirationTime time.Time
	// The ID of the event that has been scheduled to expire/cancel the prediction.
	ExpirationEventId JacuzziId
	// The possible outcomes that folks can bet on. Externally identified via letter IDs (A, B, C, etc.),
	// where index 0 = 'A'.
	Outcomes []string
	// The ID of the outcome that won (e.g. 'A'). Will be 0 if the prediction isn't resolved.
	WinningOutcomeId rune
	// People's bets on the various options, indexed by user ID. Each user can only bet on one outcome.
	Bets map[string]PredictionBet
	// The ID of the message in the server that currently holds the controls for interacting with the prediction.
	// Whenever the prediction is modified, we update the text of this message as well to reflect the new state.
	MessageId string
	// The channel where the prediction message lives.
	ChannelId string

	// The inner state of the prediction. Don't touch.
	Resolved          bool
	ManuallyCancelled bool
}

type PredictionStateUpdateEventPayload struct {
	ServerId     string
	ChannelId    string
	PredictionId JacuzziId
}

type OutcomePool struct {
	// Total of all bets in the pool for this outcome
	Total float64
	// The individual bets that compose the pool, keyed by user ID
	Bets map[string]float64
}

// Get the current state of the prediction
func (this *Prediction) GetState(now time.Time) PredictionState {
	if this.ManuallyCancelled {
		return PredictionCancelled
	} else if this.Resolved {
		return PredictionResolved
	} else if !now.Before(this.ExpirationTime) {
		return PredictionCancelled
	} else if !now.Before(this.BettingCloseTime) {
		return BettingLocked
	} else {
		return BettingOpen
	}
}

// Get details on each outcome pool
func (this *Prediction) OutcomePools() []*OutcomePool {
	result := make([]*OutcomePool, len(this.Outcomes))
	for index, _ := range result {
		result[index] = &OutcomePool{
			Total: 0,
			Bets:  map[string]float64{},
		}
	}
	for userId, bet := range this.Bets {
		index := bet.OutcomeId - 'A'
		pool := result[index]
		pool.Bets[userId] = bet.Wager
		pool.Total += bet.Wager
	}
	return result
}

func (this *Prediction) StateString(now time.Time) string {
	state := this.GetState(now)
	switch state {
	case BettingOpen:
		return "Active and accepting bets"
	case BettingLocked:
		return "Active and no longer accepting bets"
	case PredictionResolved:
		return "Resolved and paid out"
	case PredictionCancelled:
		return "Expired/cancelled and refunded"
	}
	panic(fmt.Errorf("Unknown prediction state for ID #%d: %d", this.Id, state))
}

func (this *Prediction) PossibleGain(bet PredictionBet, outcomePools []*OutcomePool) float64 {
	theirPool := outcomePools[bet.OutcomeId-'A']
	theirShare := bet.Wager / theirPool.Total
	var otherPoolsTotal float64
	for otherOutcomeId, otherPool := range outcomePools {
		if rune(otherOutcomeId+'A') != bet.OutcomeId {
			otherPoolsTotal += otherPool.Total
		}
	}
	return theirShare * otherPoolsTotal
}

func (this *Prediction) DisplayString(now time.Time) string {
	state := this.GetState(now)
	message := fmt.Sprintf("# Prediction - %s", this.Question)
	message += fmt.Sprintf("\n- **%s**", this.StateString(now))
	message += fmt.Sprintf("\n- Created by <@%s> at `%s`", this.Creator, this.CreationTime.Format(utils.TimeFormat))
	if now.After(this.BettingCloseTime) {
		message += fmt.Sprintf("\n- Betting closed at `%s`", this.BettingCloseTime.Format(utils.TimeFormat))
	} else {
		message += fmt.Sprintf("\n- Betting is open until `%s`", this.BettingCloseTime.Format(utils.TimeFormat))
	}
	if now.After(this.ExpirationTime) {
		message += fmt.Sprintf("\n- Expired at `%s`", this.ExpirationTime.Format(utils.TimeFormat))
	} else {
		message += fmt.Sprintf("\n- Expires at `%s`", this.ExpirationTime.Format(utils.TimeFormat))
	}

	message += "\n## Possible outcomes"
	outcomePools := this.OutcomePools()
	for index, outcome := range this.Outcomes {
		id := rune('A' + index)
		if state == PredictionResolved {
			if this.WinningOutcomeId == id {
				message += fmt.Sprintf("\n### [:tada: WINNER :tada:] %c. %s", id, outcome)
			} else {
				message += fmt.Sprintf("\n### [LOSER]  %c. %s", id, outcome)
			}
		} else {
			message += fmt.Sprintf("\n### %c. %s", id, outcome)
		}
		totalPool := outcomePools[index].Total
		message += fmt.Sprintf("\n- Total pool: %s point%s",
			utils.FormatUIFloat(totalPool), utils.Plural(totalPool))
		if len(outcomePools[index].Bets) > 0 {
			message += "\n- Individual bets:"
			for userId, wager := range outcomePools[index].Bets {
				possibleGain := this.PossibleGain(this.Bets[userId], outcomePools)
				possibleGainPercent := possibleGain / wager * 100
				message += fmt.Sprintf("\n  - <@%s> bet %s point%s",
					userId, utils.FormatUIFloat(wager), utils.Plural(wager),
				)
				switch state {
				case PredictionResolved:
					if this.WinningOutcomeId == id {
						message += fmt.Sprintf(", and WON: +%s (+%s%%)",
							utils.FormatUIFloat(possibleGain), utils.FormatUIFloat(possibleGainPercent))
					} else {
						message += ", and LOST"
					}
				case PredictionCancelled:
					message += " (refunded)"
				default:
					message += fmt.Sprintf(". Possible winnings: +%s (+%s%%)",
						utils.FormatUIFloat(possibleGain), utils.FormatUIFloat(possibleGainPercent))
				}
			}
		}
	}
	return message
}

func (this *Prediction) GetOutcome(id rune) string {
	return this.Outcomes[id-'A']
}
