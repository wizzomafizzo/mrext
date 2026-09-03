import React, {useEffect, useMemo} from "react";
import {useMutation} from "@tanstack/react-query";

import Button from "@mui/material/Button";

import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import ListItemButton from "@mui/material/ListItemButton";
import TextField from "@mui/material/TextField";
import Stack from "@mui/material/Stack";
import LinearProgress from "@mui/material/LinearProgress";
import Typography from "@mui/material/Typography";
import Grid from "@mui/material/Grid";
import SearchIcon from "@mui/icons-material/Search";
import {ControlApi} from "../lib/api";
import {useIndexedSystems} from "../lib/queries";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import {Game, SearchResults} from "../lib/models";
import IconButton from "@mui/material/IconButton";
import ScrollToTopFab from "./ScrollToTop";
import {useServerStateStore} from "../lib/store";
import useWs from "./WebSocket";
import Box from "@mui/material/Box";
import FormControl from "@mui/material/FormControl";
import CircularProgress from "@mui/material/CircularProgress";
import Dialog from "@mui/material/Dialog";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import ShortcutIcon from "@mui/icons-material/Shortcut";
import QrCode2Icon from "@mui/icons-material/QrCode2"
import {SingleShortcut} from "./Shortcuts";
import SyncIcon from "@mui/icons-material/Sync";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogActions from "@mui/material/DialogActions";
import TapAndPlayIcon from '@mui/icons-material/TapAndPlay';
import {getApiEndpoint} from '../lib/api';
import QRCodeStyling from 'qr-code-styling';

function downloadQrCode(options: {
  path: string;
  name: string;
}) {
  const {path, name} = options;
  const data = `${getApiEndpoint()}/l/${btoa(path).replace(/=+$/, '')}`;
  const qr = new QRCodeStyling({
    data,
    type: 'canvas',
    width: 1024,
    height: 1024,
    qrOptions: {
      errorCorrectionLevel: 'L',
      mode: 'Byte',
    },
  });
  return qr._canvasDrawingPromise?.then(() => {
    qr._canvas?.toBlob((blob) => {
      if (!blob) return;
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `${name}.png`;
      a.click();
    });
  });
}

function SearchResultsList(props: {
  results?: SearchResults;
  selectedGame: Game | null;
  setSelectedGame: (game: Game | null) => void;
  setOpenShortcut: (open: boolean) => void;
}) {
  const api = new ControlApi();
  const launchGame = useMutation({
    mutationFn: api.launchGame,
  });

  const displayed = new Set<string>();
  const displayResults: Game[] = [];
  const [gameInfoOpen, setGameInfoOpen] = React.useState(false);

  const [nfcRunning, setNfcRunning] = React.useState(false);
  const [waitingNfc, setWaitingNfc] = React.useState(false);

  useEffect(() => {
    api.nfcStatus().then((status) => {
      setNfcRunning(status.running);
    });
  }, []);

  if (props.results && props.results.data) {
    for (const game of props.results.data) {
      const gameId = game.system.id + "/" + game.name;

      if (displayed.has(gameId)) {
        continue;
      }

      displayed.add(gameId);
      displayResults.push(game);
    }
  }

  return (
    <Box width={1}>
      {props.results ? (
        <Typography
          variant="body2"
          sx={{marginTop: "0.5em", textAlign: "center"}}
        >
          Found {displayed.size} {displayed.size === 1 ? "game" : "games"}
        </Typography>
      ) : null}
      <List sx={{marginTop: 2}} disablePadding>
        {displayResults
          ?.sort((a, b) => a.system.id.localeCompare(b.system.id))
          .map((game) => (
            <ListItem key={game.path} disableGutters disablePadding>
              <ListItemButton
                onClick={() => {
                  props.setSelectedGame(game);
                  setGameInfoOpen(true);
                }}
                disableGutters
              >
                <ListItemText
                  primary={game.name}
                  secondary={game.system.name}
                />
              </ListItemButton>
            </ListItem>
          ))}
      </List>
      <Dialog
        open={gameInfoOpen}
        onClose={() => {
          setGameInfoOpen(false);
          props.setSelectedGame(null);
        }}
      >
        {props.selectedGame ? (
          <Stack sx={{m: 1, minWidth: "250px"}}>
            <Box sx={{mb: 2}}>
              <Typography>{props.selectedGame.name}</Typography>
              <Typography variant="body2">
                {props.selectedGame.system.name}
              </Typography>
            </Box>
            <Button
              variant="contained"
              onClick={() => {
                if (props.selectedGame) {
                  launchGame.mutate(props.selectedGame.path);
                  setGameInfoOpen(false);
                }
              }}
              startIcon={<PlayArrowIcon/>}
            >
              Launch
            </Button>
            <Button
              variant="outlined"
              sx={{mt: 1}}
              startIcon={<ShortcutIcon/>}
              onClick={() => props.setOpenShortcut(true)}
            >
              Create shortcut
            </Button>
            {nfcRunning ? (
              <Button
                variant="outlined"
                sx={{mt: 1}}
                startIcon={<TapAndPlayIcon/>}
                onClick={() => {
                  if (props.selectedGame) {
                    setWaitingNfc(true);
                    api.nfcWrite({
                      path: props.selectedGame.path
                    }).then(() => {
                      setGameInfoOpen(false);
                    }).finally(() => {
                      setWaitingNfc(false);
                    });
                  }
                }}
              >
                {waitingNfc ? "Waiting for tag..." : "Write to NFC tag"}
              </Button>
            ) : null}
            <Button
              variant="outlined"
              sx={{mt: 1}}
              startIcon={<QrCode2Icon/>}
              onClick={() => props.selectedGame && downloadQrCode({
                path: props.selectedGame.path,
                name: props.selectedGame.name,
              })}
            >
              Create QR-Code
            </Button>
            <Button
              sx={{mt: 1}}
              onClick={() => {
                setGameInfoOpen(false);
                props.setSelectedGame(null);
              }}
            >
              Close
            </Button>
          </Stack>
        ) : null}
      </Dialog>
    </Box>
  );
}

export default function Search() {
  const api = new ControlApi();
  const systems = useIndexedSystems();
  const serverState = useServerStateStore();
  const ws = useWs();

  const [searchQuery, setSearchQuery] = React.useState("");
  const [searchSystem, setSearchSystem] = React.useState("all");

  const [openShortcut, setOpenShortcut] = React.useState(false);
  const [selectedGame, setSelectedGame] = React.useState<Game | null>(null);

  const [regenOpen, setRegenOpen] = React.useState(false);

  useEffect(() => {
    ws.sendMessage("getIndexStatus");
  }, []);

  const searchGames = useMutation({
    mutationFn: () =>
      api.searchGames({
        query: searchQuery,
        system: searchSystem,
      }),
  });

  const startIndex = useMutation({
    mutationFn: api.startSearchIndex,
  });

  const resultsList = useMemo(() => {
    return (
      <SearchResultsList
        results={searchGames.data}
        selectedGame={selectedGame}
        setSelectedGame={setSelectedGame}
        setOpenShortcut={setOpenShortcut}
      />
    );
  }, [searchGames.data, selectedGame]);

  if (ws.readyState !== WebSocket.OPEN) {
    return <></>;
  }

  if (serverState.search.indexing) {
    return (
      <Box m={2}>
        <Typography variant="h5">Indexing games...</Typography>
        <LinearProgress
          variant="determinate"
          value={
            serverState.search.currentStep === 0 ||
            serverState.search.totalSteps === 0
              ? 0
              : (serverState.search.currentStep /
                serverState.search.totalSteps) *
              100
          }
        />
        <Typography>{serverState.search.currentDesc}</Typography>
      </Box>
    );
  }

  if (!serverState.search.ready) {
    return (
      <Box m={2} sx={{textAlign: "center"}}>
        <Typography sx={{marginBottom: 2}}>
          Searching needs an index of game files to be created. This is only
          required on first setup, or if the games on disk have changed.
        </Typography>
        <Button
          variant="contained"
          onClick={() => {
            startIndex.mutate();
          }}
        >
          Generate Index
        </Button>
      </Box>
    );
  }

  if (openShortcut && selectedGame) {
    return (
      <SingleShortcut
        path={selectedGame?.path}
        back={() => setOpenShortcut(false)}
      />
    );
  }

  return (
    <Box m={2}>
      <Grid
        container
        sx={{alignItems: "center"}}
        spacing={{xs: 2, sm: 2, md: 3}}
        columns={{xs: 4, sm: 12, md: 12}}
      >
        <Grid item xs={8}>
          <FormControl fullWidth>
            <TextField
              variant="outlined"
              autoComplete="off"
              sx={{width: "100%"}}
              onChange={(e) => setSearchQuery(e.target.value)}
              onKeyDown={(e) => {
                if (e.key === "Enter") {
                  searchGames.mutate();
                }
              }}
              placeholder="Search"
            />
          </FormControl>
          <FormControl fullWidth sx={{marginTop: 2}}>
            <Stack width={1} direction="row">
              <Select
                value={searchSystem}
                onChange={(e) => setSearchSystem(e.target.value)}
                size="small"
                sx={{flexGrow: 1}}
              >
                <MenuItem value="all">All systems</MenuItem>
                {systems.data?.systems
                  .sort((a, b) => a.name.localeCompare(b.name))
                  .map((system) => (
                    <MenuItem key={system.id} value={system.id}>
                      {system.name}
                    </MenuItem>
                  ))}
              </Select>
              <Box>
                <IconButton onClick={() => setRegenOpen(true)}>
                  <SyncIcon/>
                </IconButton>
              </Box>
            </Stack>
          </FormControl>
        </Grid>
        <Grid item xs={4}>
          <Button
            variant="contained"
            sx={{width: "100%"}}
            onClick={() => {
              searchGames.mutate();
            }}
            startIcon={<SearchIcon/>}
          >
            Search
          </Button>
        </Grid>
      </Grid>
      <Stack sx={{alignItems: "center"}}>
        {searchGames.isLoading ? (
          <div style={{textAlign: "center", marginTop: "2em"}}>
            <CircularProgress/>
          </div>
        ) : (
          resultsList
        )}
        <ScrollToTopFab/>
      </Stack>

      <Dialog open={regenOpen} onClose={() => setRegenOpen(false)}>
        <DialogContent>
          <DialogContentText>Regenerate the search index?</DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setRegenOpen(false)}>Cancel</Button>
          <Button
            onClick={() => {
              setRegenOpen(false);
              startIndex.mutate();
            }}
            variant="contained"
          >
            Regenerate
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}
