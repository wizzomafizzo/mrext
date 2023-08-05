package config

const UserConfigEnv = "MREXT_CONFIG"
const UserAppPathEnv = "MREXT_APP_PATH"

const ActiveGameFile = TempFolder + "/ACTIVEGAME"
const SearchDbFile = SdFolder + "/search.db"
const PlayLogDbFile = SdFolder + "/playlog.db"

const PidFileTemplate = TempFolder + "/%s.pid"
const LogFileTemplate = TempFolder + "/%s.log"

const ScriptsConfigFolder = ScriptsFolder + "/.config"
const MrextConfigFolder = ScriptsConfigFolder + "/mrext"

const ArcadeDBUrl = "https://api.github.com/repositories/521644036/contents/ArcadeDatabase_CSV"
const ArcadeDBFile = MrextConfigFolder + "/ArcadeDatabase.csv"
