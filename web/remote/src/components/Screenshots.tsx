import React from "react";
import { useQuery, useMutation } from "@tanstack/react-query";

import Fab from "@mui/material/Fab";
import { CircularProgress } from "@mui/material";

import Card from "@mui/material/Card";
import CardActions from "@mui/material/CardActions";
import CardContent from "@mui/material/CardContent";
import CardMedia from "@mui/material/CardMedia";
import Grid from "@mui/material/Grid";
import Stack from "@mui/material/Stack";

import IconButton from "@mui/material/IconButton";
import Button from "@mui/material/Button";
import Typography from "@mui/material/Typography";

import Dialog from "@mui/material/Dialog";
import DialogActions from "@mui/material/DialogActions";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";

import AddAPhotoIcon from "@mui/icons-material/AddAPhoto";
import DeleteIcon from "@mui/icons-material/Delete";
import OpenInNewIcon from "@mui/icons-material/OpenInNew";

import { ControlApi } from "../lib/api";

export default function Screenshots() {
  const api = new ControlApi();

  const allScreenshots = useQuery({
    queryKey: ["screenshots"],
    queryFn: api.getScreenshots,
  });

  const takeScreenshot = useMutation({
    mutationFn: api.takeScreenshot,
    onSuccess: () => {
      allScreenshots.refetch();
    },
  });

  const deleteScreenshot = useMutation({
    mutationFn: api.deleteScreenshot,
    onSuccess: () => {
      allScreenshots.refetch();
    },
  });

  const [openDelete, setOpenDelete] = React.useState(false);
  const [deleteId, setDeleteId] = React.useState("");

  return (
    <div style={{ margin: "15px" }}>
      <Fab
        color="primary"
        style={{ position: "fixed", bottom: 16, right: 16 }}
        onClick={() => takeScreenshot.mutate()}
      >
        {takeScreenshot.isLoading ? (
          <CircularProgress color="inherit" />
        ) : (
          <AddAPhotoIcon />
        )}
      </Fab>

      <Grid
        container
        spacing={{ xs: 2, md: 3 }}
        columns={{ xs: 4, sm: 8, md: 12 }}
      >
        {allScreenshots.data
          ?.slice()
          .sort(
            (a, b) =>
              new Date(b.modified).getTime() - new Date(a.modified).getTime()
          )
          .map((screenshot) => (
            <Grid item xs={4} sm={4} md={4} key={screenshot.path}>
              <Card>
                <CardMedia
                  component="img"
                  image={api.getScreenshotUrl(screenshot.path)}
                  alt={screenshot.path}
                  loading="lazy"
                />
                <CardContent sx={{ paddingBottom: 0 }}>
                  <Typography gutterBottom component="div">
                    {screenshot.game}
                  </Typography>
                  <Typography variant="body2" color="text.secondary">
                    Added: {new Date(screenshot.modified).toLocaleString()}
                    <br />
                    Core: {screenshot.core}
                  </Typography>
                </CardContent>
                <CardActions>
                  <Stack
                    sx={{ width: "100%" }}
                    direction="row"
                    spacing={1}
                    justifyContent="space-between"
                  >
                    {/* <Button size="small">
                                        <ShareIcon fontSize="small" /> Share
                                    </Button> */}
                    <Button
                      color="error"
                      size="small"
                      onClick={() => {
                        setDeleteId(screenshot.path);
                        setOpenDelete(true);
                      }}
                    >
                      <DeleteIcon fontSize="small" /> Delete
                    </Button>
                    <a
                      href={api.getScreenshotUrl(screenshot.path)}
                      target="_blank"
                      rel="noreferrer"
                    >
                      <IconButton>
                        <OpenInNewIcon fontSize="small" />
                      </IconButton>
                    </a>
                  </Stack>
                </CardActions>
              </Card>
            </Grid>
          ))}
      </Grid>

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
            Delete this screenshot?
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
              deleteScreenshot.mutate(deleteId);
            }}
          >
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  );
}
