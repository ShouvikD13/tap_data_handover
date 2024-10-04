package initializers

type MainContainer struct {
	UtilContainer              *UtilContainer
	ClientContainer            *ClientContainer
	ClientGlobalValueContainer *ClientGlobalValueContainer
	LogOnContainer             *LogOnContainer
	LogOnGlobalValueContainer  *LogOnGlobalValueContainer
	LogOffContainer            *LogOffContainer
	LogOffGlobalValueContainer *LogOffGlobalValueContainer
	ESRContainer               *ESRContainer
	ESRGlobalValueContainer    *ESRGlobalValueContainer
}
