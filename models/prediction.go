package models

import (
	"fmt"
	"time"

	"github.com/janimationd/JacuzziBot/utils"
)

type PredictionState int

const (
	AcceptingBets PredictionState = iota
	Locked
	Resolved
	Cancelled
)

type PredictionBet struct {
	OutcomeId rune
	Wager     float64
}

type Prediction struct {
	// The unique ID of the prediction
	Id JacuzziId
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
	// People's bets on the various options, indexed by user ID. Each user can only bet on one outcome.
	Bets map[string]PredictionBet
	// The current state of the prediction.
	State PredictionState
	// The ID of the message in the server that currently holds the controls for interacting with the prediction.
	// Whenever the prediction is modified, we update the text of this message as well to reflect the new state.
	MessageId string
}

type OutcomePool struct {
	// Total of all bets in the pool for this outcome
	Total float64
	// The individual bets that compose the pool, keyed by user ID
	Bets map[string]float64
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

func (this *Prediction) StateString() string {
	switch this.State {
	case AcceptingBets:
		return "Active and accepting bets"
	case Locked:
		return "Active and no longer accepting bets"
	case Resolved:
		return "Resolved and paid out"
	case Cancelled:
		return "Expired/cancelled and refunded"
	}
	panic(fmt.Errorf("Unknown prediction state for ID #%d: %d", this.Id, this.State))
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

func (this *Prediction) DisplayString() string {
	message := fmt.Sprintf("# Prediction #%d - %s", this.Id, this.Question)
	message += fmt.Sprintf("\n- **%s**", this.StateString())
	message += fmt.Sprintf("\n- Created at `%s`", this.CreationTime.Format(utils.TimeFormat))
	now := time.Now()
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
		id := 'A' + index
		message += fmt.Sprintf("\n### %c. %s", id, outcome)
		totalPool := outcomePools[index].Total
		message += fmt.Sprintf("\n- Total pool: %s point%s",
			utils.FormatUIFloat(totalPool), utils.Plural(totalPool))
		if len(outcomePools[index].Bets) > 0 {
			message += "\n- Individual bets:"
			for userId, wager := range outcomePools[index].Bets {
				possibleGain := this.PossibleGain(this.Bets[userId], outcomePools)
				possibleGainPercent := possibleGain / wager * 100
				message += fmt.Sprintf("\n  - <@%s> bet %s point%s. Possible winnings: +%s more (+%s%%).",
					userId, utils.FormatUIFloat(wager), utils.Plural(wager), utils.FormatUIFloat(possibleGain),
					utils.FormatUIFloat(possibleGainPercent),
				)
			}
		}
	}
	return message
}
