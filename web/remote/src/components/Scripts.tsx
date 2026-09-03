import { ControlApi } from "../lib/api";
import { useServerStateStore, useUIStateStore } from "../lib/store";
import { useQuery } from "@tanstack/react-query";
import Box from "@mui/material/Box";
import { useNavigate } from "react-router-dom";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import StarIcon from "@mui/icons-material/Star";
import StarBorderIcon from "@mui/icons-material/StarBorder";
import Grid from "@mui/material/Grid";
import { Script } from "../lib/models";
import { useScriptsList } from "../lib/queries";
import { useEffect } from "react";
import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";

export function Scripts() {
  const api = new ControlApi();
  const uiState = useUIStateStore();
  const allScripts = useScriptsList();
  const navigate = useNavigate();
  const serverState = useServerStateStore();

  useEffect(() => {
    allScripts.refetch().catch((err) => {
      console.log(err);
    });
  }, [serverState.activeCore]);

  const favorites = uiState.favoriteScripts;
  const isFavorite = (filename: string) => {
    return favorites.includes(filename);
  };
  const addFavorite = (filename: string) => {
    if (!isFavorite(filename)) {
      uiState.setFavoriteScripts([...favorites, filename]);
    }
  };
  const removeFavorite = (filename: string) => {
    uiState.setFavoriteScripts(favorites.filter((f) => f !== filename));
  };

  allScripts.data?.scripts.sort((a, b) => a.name.localeCompare(b.name));

  const displayFavorites: Script[] = [];
  const displayRest: Script[] = [];

  allScripts.data?.scripts.forEach((script) => {
    if (isFavorite(script.filename)) {
      displayFavorites.push(script);
    } else {
      displayRest.push(script);
    }
  });

  return (
    <Box m={2} mb={0}>
      <Grid container spacing={1}>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            size="small"
            fullWidth
            onClick={() => {
              api
                .openConsole()
                .catch((err) => {
                  console.log(err);
                })
                .then(() => {
                  navigate("/control");
                });
            }}
            disabled={serverState.activeCore !== ""}
          >
            Launch console
          </Button>
        </Grid>
        <Grid item xs={6}>
          <Button
            variant="outlined"
            size="small"
            fullWidth
            onClick={() => {
              api.closeConsole().catch((err) => {
                console.log(err);
              });
            }}
            disabled={serverState.activeCore !== ""}
          >
            Exit console
          </Button>
        </Grid>
      </Grid>
      <List>
        {displayFavorites.map((script) => {
          return (
            <ListItem
              disablePadding
              key={script.filename}
              secondaryAction={
                <IconButton
                  sx={{ p: 0.5, pt: 1, mr: -2.5 }}
                  onClick={() => {
                    removeFavorite(script.filename);
                  }}
                >
                  <StarIcon />
                </IconButton>
              }
              sx={{ pt: 0.5 }}
            >
              <ListItemButton
                disableGutters
                onClick={() => {
                  api
                    .runScript(script.filename)
                    .catch((err) => {
                      console.log(err);
                    })
                    .then(() => {
                      navigate("/control");
                    });
                }}
                disabled={!allScripts.data?.canLaunch}
              >
                {script.name}
              </ListItemButton>
            </ListItem>
          );
        })}
        {displayRest.map((script) => {
          return (
            <ListItem
              disablePadding
              key={script.filename}
              secondaryAction={
                <IconButton
                  sx={{ p: 0.5, pt: 1, mr: -2.5 }}
                  onClick={() => {
                    addFavorite(script.filename);
                  }}
                >
                  <StarBorderIcon />
                </IconButton>
              }
              sx={{ pt: 0.5 }}
            >
              <ListItemButton
                disableGutters
                onClick={() => {
                  api
                    .runScript(script.filename)
                    .catch((err) => {
                      console.log(err);
                    })
                    .then(() => {
                      navigate("/control");
                    });
                }}
                disabled={!allScripts.data?.canLaunch}
              >
                {script.name}
              </ListItemButton>
            </ListItem>
          );
        })}
      </List>
    </Box>
  );
}
