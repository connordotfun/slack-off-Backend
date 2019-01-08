package elo

import "math"

// CalculateNewRatings calculates the new ratings for the winner and loser of a matchup
func CalculateNewRatings(winnerRating float64, loserRating float64) (float64, float64) {
	ratingDifference := loserRating - winnerRating

	expectedScoreForWinner := calculateExpectedScore(ratingDifference)
	expectedScoreForLoser := calculateExpectedScore(-ratingDifference)

	newWinnerRating := calculateNewElo(winnerRating, expectedScoreForWinner, 1)
	newLoserRating := calculateNewElo(loserRating, expectedScoreForLoser, 0)

	return newWinnerRating, newLoserRating
}

func calculateExpectedScore(ratingDifference float64) float64 {
	return 1 / (1 + math.Pow(10, (ratingDifference/400)))
}

func calculateNewElo(baseRating float64, expectedScore float64, actualScore float64) float64 {
	return baseRating + 8*(actualScore-expectedScore)
}
