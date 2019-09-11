package main

type SteamGameSummaryStatistics struct {
	AverageDecline        int
	AverageGain           int
	AverageMaxPlayerCount int
	AverageMinPlayerCount int
	AveragePlayerCount    int
	MonthsSinceRelease    int
	PeakPlayers           int
	PeakPlayersDate       string
	TroughPlayers         int
	TroughPlayersDate     string
	YearsSinceRelease     int
}

func NewSteamGameSummaryStatistics(s *SteamChartPage) SteamGameSummaryStatistics {
	var (
		averageDecline        int
		averageGain           int
		averageMaxPlayerCount int
		averageMinPlayerCount int
		averagePlayerCount    int
		monthsSinceRelease    int
		peakPlayers           int
		peakPlayersDate       string
		troughPlayersDate     string
		yearsSinceRelease     int

		troughPlayers = int(^uint(0) >> 1)
	)
	if len(s.Growth) > 0 {
		for _, s := range s.Growth {
			if s.Gain > 0 {
				averageGain = averageGain + int(s.Gain)
				averageMaxPlayerCount = averageMaxPlayerCount + s.PlayersPeak
			} else {
				averageDecline = averageDecline + int(s.Gain)
				averageMinPlayerCount = averageMinPlayerCount + s.PlayersPeak
			}
			if s.PlayersPeak > peakPlayers {
				peakPlayers = s.PlayersPeak
				peakPlayersDate = s.Month
			}
			if s.PlayersPeak < troughPlayers {
				troughPlayers = s.PlayersPeak
				troughPlayersDate = s.Month
			}
			averagePlayerCount = averagePlayerCount + int(s.PlayersAverage)
			monthsSinceRelease = monthsSinceRelease + 1
		}
		averageDecline = averageDecline / monthsSinceRelease
		averageGain = averageGain / monthsSinceRelease
		averageMaxPlayerCount = averageMaxPlayerCount / monthsSinceRelease
		averageMinPlayerCount = averageMinPlayerCount / monthsSinceRelease
		averagePlayerCount = averagePlayerCount / monthsSinceRelease
		yearsSinceRelease = monthsSinceRelease / 12
	}
	return SteamGameSummaryStatistics{
		AverageDecline:        averageDecline,
		AverageGain:           averageGain,
		AverageMaxPlayerCount: averageMaxPlayerCount,
		AverageMinPlayerCount: averageMinPlayerCount,
		AveragePlayerCount:    averagePlayerCount,
		MonthsSinceRelease:    monthsSinceRelease,
		PeakPlayers:           peakPlayers,
		PeakPlayersDate:       peakPlayersDate,
		TroughPlayers:         troughPlayers,
		TroughPlayersDate:     troughPlayersDate,
		YearsSinceRelease:     yearsSinceRelease}
}
