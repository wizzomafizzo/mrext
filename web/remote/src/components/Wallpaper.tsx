import React from "react";
import { useQuery, useMutation } from "@tanstack/react-query";

import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";

import Button from "@mui/material/Button";
import IconButton from "@mui/material/IconButton";
import Typography from "@mui/material/Typography";

import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import ClearIcon from "@mui/icons-material/Clear";
import StarIcon from "@mui/icons-material/Star";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";

import { ControlApi } from "../lib/api";
import ScrollToTopFab from "./ScrollToTop";
import Box from "@mui/material/Box";
import Paper from "@mui/material/Paper";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";

export default function Wallpaper() {
  const api = new ControlApi();

  const allWallpapers = useQuery({
    queryKey: ["wallpaper"],
    queryFn: api.getWallpapers,
    onSuccess: (wp) => {
      setBackgroundMode(wp.backgroundMode);
    },
  });

  const deleteWallpaper = useMutation({
    mutationFn: api.deleteWallpaper,
    onSuccess: () => {
      allWallpapers.refetch();
    },
  });

  const uploadWallpaper = useMutation({
    mutationFn: api.getWallpapers,
    onSuccess: () => {
      allWallpapers.refetch();
    },
  });

  const activateWallpaper = useMutation({
    mutationFn: api.setWallpaper,
    onSuccess: () => {
      allWallpapers.refetch();
    },
  });

  const [openDelete, setOpenDelete] = React.useState(false);
  const [deleteId, setDeleteId] = React.useState("");

  const [backgroundMode, setBackgroundMode] = React.useState(2);

  return (
    <>
      <Paper
        sx={{
          boxShadow: 2,
          width: 1,
          height: "55px",
          zIndex: 1,
          borderRadius: 0,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          padding: 1,
        }}
      >
        <Select
          size="small"
          value={backgroundMode}
          onChange={(e) => {
            const v = parseInt(e.target.value as string, 10);
            setBackgroundMode(v);
            api.setMenuBackgroundMode({ mode: v }).then(() => {
              allWallpapers.refetch();
            });
          }}
        >
          <MenuItem value={0}>None (static)</MenuItem>
          <MenuItem value={2}>Wallpaper</MenuItem>
          <MenuItem value={4}>Horizontal bars #1</MenuItem>
          <MenuItem value={6}>Horizontal bars #2</MenuItem>
          <MenuItem value={8}>Vertical bars #1</MenuItem>
          <MenuItem value={10}>Vertical bars #2</MenuItem>
          <MenuItem value={12}>Spectrum</MenuItem>
          <MenuItem value={14}>Black</MenuItem>
        </Select>
        <Button
          onClick={() =>
            api.unsetWallpaper().then(() => {
              allWallpapers.refetch();
            })
          }
          disabled={
            allWallpapers.isLoading || allWallpapers.data?.active === ""
          }
        >
          Clear wallpaper
        </Button>
      </Paper>

      <Box m={2}>
        <Grid
          container
          spacing={{ xs: 2, md: 3 }}
          columns={{ xs: 4, sm: 8, md: 12 }}
        >
          {allWallpapers.data?.wallpapers
            .slice()
            .sort()
            .map((wallpaper) => (
              <Grid item xs={4} sm={4} md={4} key={wallpaper.filename}>
                <Card
                  sx={
                    wallpaper.active
                      ? {
                          boxShadow: (theme) =>
                            `inset 0px 0px 0px 3px ${theme.palette.primary.main}`,
                          borderRadius: "5px",
                        }
                      : {}
                  }
                >
                  <CardMedia
                    component="img"
                    image={api.getWallpaperUrl(wallpaper.filename)}
                    alt={wallpaper.name}
                    loading="lazy"
                  />
                  <CardContent sx={{ paddingBottom: 0 }}>
                    <Typography gutterBottom component="div">
                      {wallpaper.name}
                    </Typography>
                    <Typography
                      variant="body2"
                      color="text.secondary"
                    ></Typography>
                  </CardContent>
                  <CardActions>
                    <Stack
                      sx={{ width: "100%" }}
                      direction="row"
                      spacing={1}
                      justifyContent="space-between"
                    >
                      {wallpaper.active ? (
                        <Button
                          size="small"
                          onClick={() =>
                            api.unsetWallpaper().then(() => {
                              allWallpapers.refetch();
                            })
                          }
                          startIcon={<ClearIcon fontSize="small" />}
                        >
                          Clear
                        </Button>
                      ) : (
                        <Button
                          size="small"
                          onClick={() => {
                            activateWallpaper.mutate(wallpaper.filename);
                          }}
                          startIcon={<StarIcon fontSize="small" />}
                        >
                          Set
                        </Button>
                      )}
                      <a
                        href={api.getWallpaperUrl(wallpaper.filename)}
                        target="_blank"
                        rel="noreferrer"
                      >
                        <IconButton>
                          <OpenInNewIcon fontSize="small" />
                        </IconButton>
                      </a>
                      {/* <Button size="small">
                                        <DriveFileRenameOutlineIcon fontSize="small" />{" "}
                                        Rename
                                    </Button> */}
                      {/* <IconButton
                                            color="error"
                                            size="small"
                                            onClick={() => {
                                                setDeleteId(
                                                    screenshot.filename
                                                );
                                                setOpenDelete(true);
                                            }}
                                        >
                                            <DeleteIcon fontSize="small" />
                                        </IconButton> */}
                    </Stack>
                  </CardActions>
                </Card>
              </Grid>
            ))}
        </Grid>
      </Box>

      <ScrollToTopFab />

      <Dialog
        open={openDelete}
        onClose={() => {
          setOpenDelete(false);
          setDeleteId("");
        }}
        aria-labelledby="alert-dialog-title"
        aria-describedby="alert-dialog-description"
      >
        <DialogContent>
          <DialogContentText id="alert-dialog-description">
            Delete this wallpaper?
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={() => setOpenDelete(false)} autoFocus>
            Cancel
          </Button>
          <Button
            color="error"
            onClick={() => {
              setOpenDelete(false);
              deleteWallpaper.mutate(deleteId);
            }}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}
