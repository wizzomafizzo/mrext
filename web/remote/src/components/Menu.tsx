import React, {useEffect, useLayoutEffect, useState} from "react";
import {useListGamesFolder, useListMenuFolder} from "../lib/queries";
import Box from "@mui/material/Box";
import List from "@mui/material/List";
import ListItemIcon from "@mui/material/ListItemIcon";
import FolderOpenIcon from "@mui/icons-material/FolderOpen";
import DeveloperBoardIcon from "@mui/icons-material/DeveloperBoard";
import VideogameAssetIcon from "@mui/icons-material/VideogameAsset";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import ListItemText from "@mui/material/ListItemText";
import {Game, MenuItem as MEMenuItem} from "../lib/models";
import moment from "moment";
import ListItemButton from "@mui/material/ListItemButton";
import ScrollToTopFab from "./ScrollToTop";
import Typography from "@mui/material/Typography";
import LinearProgress from "@mui/material/LinearProgress";
import {ControlApi} from "../lib/api";
import Stack from "@mui/material/Stack";
import HomeIcon from "@mui/icons-material/Home";
import MenuItem from "@mui/material/MenuItem";
import MoreVertIcon from "@mui/icons-material/MoreVert";
import RadioButtonUncheckedIcon from "@mui/icons-material/RadioButtonUnchecked";
import RadioButtonCheckedIcon from "@mui/icons-material/RadioButtonChecked";
import PlayArrowIcon from '@mui/icons-material/PlayArrow';
import {
  CreateFolder,
  formatCurrentPath,
  isValidFilename,
  MenuFolderPicker, SingleShortcut,
} from "./Shortcuts";
import Paper from "@mui/material/Paper";
import IconButton from "@mui/material/IconButton";
import Popper from "@mui/material/Popper";
import Grow from "@mui/material/Grow";
import ClickAwayListener from "@mui/material/ClickAwayListener";
import MenuList from "@mui/material/MenuList";
import SortIcon from "@mui/icons-material/Sort";
import ListItem from "@mui/material/ListItem";
import Dialog from "@mui/material/Dialog";
import {DialogTitle} from "@mui/material";
import DialogContent from "@mui/material/DialogContent";
import DialogActions from "@mui/material/DialogActions";
import Button from "@mui/material/Button";
import TextField from "@mui/material/TextField";
import FormControl from "@mui/material/FormControl";
import DriveFileRenameOutlineIcon from "@mui/icons-material/DriveFileRenameOutline";
import DriveFileMoveIcon from "@mui/icons-material/DriveFileMove";
import DeleteForeverIcon from "@mui/icons-material/DeleteForever";
import Checkbox from "@mui/material/Checkbox";
import FormControlLabel from "@mui/material/FormControlLabel";
import FolderZipIcon from '@mui/icons-material/FolderZip';
import TapAndPlayIcon from "@mui/icons-material/TapAndPlay";
import ShortcutIcon from "@mui/icons-material/Shortcut";

enum Sort {
  NameAsc,
  NameDesc,
  DateAsc,
  DateDesc,
}

enum EditMode {
  None,
  Rename,
  Move,
  Delete,
}

function RenameFile(props: {
  item: MEMenuItem;
  parentContents: MEMenuItem[];
  close: () => void;
}) {
  let editName = props.item.filename;
  if (props.item.type === "folder" && editName.startsWith("_")) {
    editName = editName.substring(1);
  } else {
    editName = editName.substring(0, editName.lastIndexOf("."));
  }

  const api = new ControlApi();
  const [name, setName] = useState<string>(editName);

  return (
    <>
      <DialogContent>
        <FormControl sx={{pt: 1}}>
          <TextField
            value={name}
            onChange={(e) => setName(e.target.value)}
            label="Name"
            inputProps={{maxLength: 255}}
            autoFocus
            helperText={
              props.item.namesTxt
                ? "This name is set via the names.txt file."
                : ""
            }
          />
        </FormControl>
      </DialogContent>
      <DialogActions>
        <Button onClick={props.close}>Cancel</Button>
        <Button
          onClick={() => {
            let oldName = props.item.filename;
            let newName = name.trim() + props.item.extension;
            if (props.item.type === "folder") {
              newName = "_" + newName;
            }

            api
              .renameMenuFile({
                fromPath: props.item.parent + "/" + oldName,
                toPath: props.item.parent + "/" + newName,
              })
              .catch((err) => {
                console.log(err);
              })
              .then(() => {
                setName("");
                props.close();
              });
          }}
          disabled={!isValidFilename(name, props.parentContents)}
          variant="contained"
        >
          Rename
        </Button>
      </DialogActions>
    </>
  );
}

function DeleteFile(props: { item: MEMenuItem; close: () => void }) {
  const api = new ControlApi();
  const [deleteFolder, setDeleteFolder] = useState<boolean>(
    props.item.type !== "folder"
  );

  return (
    <>
      <DialogContent>
        <Typography>
          Are you sure you want to permanently delete
          <Box component="span" fontWeight="fontWeightMedium">
            {" "}
            {props.item.name}
          </Box>
          ? This cannot be undone.
        </Typography>
        {props.item.type === "folder" && (
          <FormControl sx={{pt: 1}}>
            <FormControlLabel
              control={
                <Checkbox
                  checked={deleteFolder}
                  onChange={(e) => setDeleteFolder(e.target.checked)}
                />
              }
              label="Also delete all files and folders inside this folder."
            />
          </FormControl>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={props.close}>Cancel</Button>
        <Button
          onClick={() => {
            api
              .deleteMenuFile({
                path: props.item.path,
              })
              .catch((err) => {
                console.log(err);
              })
              .then(() => {
                props.close();
              });
          }}
          variant="contained"
          color="error"
          disabled={!deleteFolder}
        >
          Delete
        </Button>
      </DialogActions>
    </>
  );
}

function EditSection(props: {
  item: MEMenuItem;
  parentContents: MEMenuItem[];
  refresh: () => void;
  setOpenShortcut?: (open: boolean) => void;
  setSelectedPath?: (path: string) => void;
  setOpen: (open: boolean) => void;
  handleClose: () => void;
}) {
  const api = new ControlApi();

  const [editMode, setEditMode] = useState<EditMode>(EditMode.None);

  const [nfcRunning, setNfcRunning] = React.useState(false);
  const [waitingNfc, setWaitingNfc] = React.useState(false);


  useEffect(() => {
    api.nfcStatus().then((status) => {
      setNfcRunning(status.running);
    });
  }, []);

  useEffect(() => {
    setEditMode(EditMode.None);
  }, []);


  switch (editMode) {
    case EditMode.Rename:
      return (
        <RenameFile
          item={props.item}
          parentContents={props.parentContents}
          close={props.handleClose}
        />
      );
    case EditMode.Move:
      return (
        <MenuFolderPicker
          path={props.item.parent}
          setPath={(path) => {
            api
              .renameMenuFile({
                fromPath: props.item.path,
                toPath: path + "/" + props.item.filename,
              })
              .catch((err) => {
                console.log(err);
              })
              .then(() => {
                props.refresh();
              });
          }}
          close={() => props.setOpen(false)}
          defaultPath={props.item.parent}
          verb={"Move to"}
        />
      );
    case EditMode.Delete:
      return <DeleteFile item={props.item} close={props.handleClose}/>;
    default:
      return (
        <Box sx={{minWidth: "250px"}}>
          <DialogTitle sx={{p: 2, pb: 0}}>
            {props.item.namesTxt ? props.item.namesTxt : props.item.name}
          </DialogTitle>
          <DialogContent sx={{p: 2}}>
            <List dense sx={{pt: 0}}>
              <ListItem disableGutters disablePadding>
                <ListItemText
                  primary="Filename"
                  secondary={props.item.filename}
                />
              </ListItem>
              <ListItem disableGutters disablePadding>
                <ListItemText
                  primary="Type"
                  secondary={
                    props.item.type === "folder"
                      ? "Folder"
                      : props.item.type.toUpperCase()
                  }
                />
              </ListItem>
              {!props.item.inZip ? (
                <ListItem disableGutters disablePadding>
                  <ListItemText
                    primary="Modified"
                    secondary={new Date(props.item.modified).toLocaleString()}
                  />
                </ListItem>
              ) : null}
              {props.item.system ? (
                <ListItem disableGutters disablePadding>
                  <ListItemText
                    primary="System"
                    secondary={props.item.system.name}
                  />
                </ListItem>
              ) : null}
            </List>
            <Stack spacing={1}>
              {props.item.type !== "folder" && props.item.type !== "zip" ? (
                <Button
                  variant="contained"
                  startIcon={<PlayArrowIcon/>}
                  onClick={() => {
                    api.launchFile(props.item.path).catch((err) => {
                      console.error(err);
                    }).then(() => {
                      props.setOpen(false);
                    });
                  }}
                >
                  Launch
                </Button>
              ) : null}
              {props.setOpenShortcut && props.item.system ? (
                <Button
                  variant="outlined"
                  sx={{mt: 1}}
                  startIcon={<ShortcutIcon/>}
                  onClick={() => {
                    if (props.setOpenShortcut) {
                      props.setSelectedPath && props.setSelectedPath(props.item.path)
                      props.setOpenShortcut(true)
                      props.setOpen(false)
                    }
                  }}
                >
                  Create shortcut
                </Button>
              ) : null}
              {nfcRunning && props.item.type !== "folder" && props.item.type !== "zip" ? (
                <Button
                  variant="outlined"
                  sx={{mt: 1}}
                  startIcon={<TapAndPlayIcon/>}
                  onClick={() => {
                    if (props.item.path) {
                      setWaitingNfc(true);
                      api.nfcWrite({
                        path: props.item.path
                      }).then(() => {
                        props.setOpen(false);
                      }).finally(() => {
                        setWaitingNfc(false);
                      });
                    }
                  }}
                >
                  {waitingNfc ? "Waiting for tag..." : "Write to NFC tag"}
                </Button>
              ) : null}
              {!props.item.inZip ? (
                <>
                  <Button
                    variant="outlined"
                    startIcon={<DriveFileRenameOutlineIcon/>}
                    onClick={() => setEditMode(EditMode.Rename)}
                  >
                    Rename
                  </Button>
                  <Button
                    variant="outlined"
                    startIcon={<DriveFileMoveIcon/>}
                    onClick={() => setEditMode(EditMode.Move)}
                  >
                    Move
                  </Button>
                  <Button
                    variant="outlined"
                    startIcon={<DeleteForeverIcon/>}
                    onClick={() => setEditMode(EditMode.Delete)}
                    color="error"
                  >
                    Delete
                  </Button>
                </>
              ) : null}
            </Stack>
          </DialogContent>
        </Box>
      );
    }
}

function EditFile(props: {
  item: MEMenuItem;
  parentContents: MEMenuItem[];
  refresh: () => void;
  setOpenShortcut?: (open: boolean) => void;
  setSelectedPath?: (path: string) => void;
}) {
  const [open, setOpen] = useState<boolean>(false);

  const handleClose = () => {
    setOpen(false);
    props.refresh();
  };

  return (
    <>
      <IconButton sx={{mr: -1}} onClick={() => setOpen(true)}>
        <MoreVertIcon/>
      </IconButton>
      {open && (<Dialog open={open} onClose={handleClose}>
        <EditSection
          {...props}
          setOpen={setOpen}
          handleClose={handleClose}
        />
      </Dialog>)}
    </>
  );
}

function SortFiles(props: { sort: Sort; setSort: (sort: Sort) => void }) {
  const [sortOpen, setSortOpen] = useState<boolean>(false);
  const sortAnchorRef = React.useRef<HTMLButtonElement>(null);

  const handleSortClose = (event: Event) => {
    if (
      sortAnchorRef.current &&
      sortAnchorRef.current.contains(event.target as HTMLElement)
    ) {
      return;
    }

    setSortOpen(false);
  };

  return (
    <>
      <IconButton ref={sortAnchorRef} onClick={() => setSortOpen(!sortOpen)}>
        <SortIcon/>
      </IconButton>
      <Popper
        sx={{
          zIndex: 2,
        }}
        open={sortOpen}
        anchorEl={sortAnchorRef.current}
        role={undefined}
        transition
        disablePortal
      >
        {({TransitionProps, placement}) => (
          <Grow
            {...TransitionProps}
            style={{
              transformOrigin:
                placement === "bottom" ? "center top" : "center bottom",
            }}
          >
            <Paper sx={{m: 1, mt: 0}}>
              <ClickAwayListener onClickAway={handleSortClose}>
                <MenuList dense>
                  {/*<MenuItem*/}
                  {/*  onClick={(e) => {*/}
                  {/*    setSortOpen(false);*/}
                  {/*  }}*/}
                  {/*>*/}
                  {/*  <ListItemIcon>*/}
                  {/*    <CheckBoxOutlineBlankOutlinedIcon />*/}
                  {/*  </ListItemIcon>*/}
                  {/*  <ListItemText>Show hidden files</ListItemText>*/}
                  {/*</MenuItem>*/}
                  <Typography sx={{pl: 1, fontWeight: 500}}>
                    Sort by
                  </Typography>
                  <MenuItem
                    onClick={() => {
                      props.setSort(Sort.NameAsc);
                      setSortOpen(false);
                    }}
                  >
                    <ListItemIcon>
                      {props.sort === Sort.NameAsc ? (
                        <RadioButtonCheckedIcon/>
                      ) : (
                        <RadioButtonUncheckedIcon/>
                      )}
                    </ListItemIcon>
                    <ListItemText>Name (ascending)</ListItemText>
                  </MenuItem>
                  <MenuItem
                    onClick={() => {
                      props.setSort(Sort.NameDesc);
                      setSortOpen(false);
                    }}
                  >
                    <ListItemIcon>
                      {props.sort === Sort.NameDesc ? (
                        <RadioButtonCheckedIcon/>
                      ) : (
                        <RadioButtonUncheckedIcon/>
                      )}
                    </ListItemIcon>
                    <ListItemText>Name (descending)</ListItemText>
                  </MenuItem>
                  <MenuItem
                    onClick={() => {
                      props.setSort(Sort.DateAsc);
                      setSortOpen(false);
                    }}
                  >
                    <ListItemIcon>
                      {props.sort === Sort.DateAsc ? (
                        <RadioButtonCheckedIcon/>
                      ) : (
                        <RadioButtonUncheckedIcon/>
                      )}
                    </ListItemIcon>
                    <ListItemText>Modified (ascending)</ListItemText>
                  </MenuItem>
                  <MenuItem
                    onClick={() => {
                      props.setSort(Sort.DateDesc);
                      setSortOpen(false);
                    }}
                  >
                    <ListItemIcon>
                      {props.sort === Sort.DateDesc ? (
                        <RadioButtonCheckedIcon/>
                      ) : (
                        <RadioButtonUncheckedIcon/>
                      )}
                    </ListItemIcon>
                    <ListItemText>Modified (descending)</ListItemText>
                  </MenuItem>
                </MenuList>
              </ClickAwayListener>
            </Paper>
          </Grow>
        )}
      </Popper>
    </>
  );
}

const sortItems = (items: MEMenuItem[] | undefined, sort: Sort, currentPath: string) => {
  if (!items) {
    return [];
  }

  const sorted = [...items];

  if (currentPath === "" || currentPath === ".") {
    sorted.sort((a, b) => a.name.localeCompare(b.name));
  } else {
    switch (sort) {
      case Sort.NameAsc:
        sorted.sort((a, b) => a.name.localeCompare(b.name));
        break;
      case Sort.NameDesc:
        sorted.sort((a, b) => b.name.localeCompare(a.name));
        break;
      case Sort.DateAsc:
        sorted.sort((a, b) => moment(a.modified).diff(moment(b.modified)));
        break;
      case Sort.DateDesc:
        sorted.sort((a, b) => moment(b.modified).diff(moment(a.modified)));
        break;
    }
  }

  const folders = sorted.filter((item) => item.type === "folder");
  const files = sorted.filter((item) => item.type !== "folder");

  return [...folders, ...files];
};

export function Menu() {
  const api = new ControlApi();

  const [currentPath, setCurrentPath] = useState<string>("");
  const [sort, setSort] = useState<Sort>(Sort.NameAsc);

  const listMenuFolder = useListMenuFolder(currentPath);
  const [sortedItems, setSortedItems] = useState<MEMenuItem[]>(listMenuFolder.data?.items || []);

  useLayoutEffect(() => {
    setSortedItems(sortItems(listMenuFolder.data?.items, sort, currentPath));
  }, [sort, currentPath, listMenuFolder.data?.items]);

  const icon = (item: MEMenuItem) => {
    switch (item.type) {
      case "folder":
        return <FolderOpenIcon/>;
      case "rbf":
        return <DeveloperBoardIcon/>;
      default:
        return <VideogameAssetIcon/>;
    }
  };

  const resetScroll = () => {
    window.scrollTo({
      top: 0,
    });
  };

  return (
    <>
      <Paper
        sx={{
          boxShadow: 2,
          position: "fixed",
          width: 1,
          height: "55px",
          zIndex: 1,
          borderRadius: 0,
        }}
      >
        <Stack
          direction="row"
          sx={{p: 1, pl: 2, alignItems: "center", height: "55px"}}
        >
          <IconButton
            sx={{pl: 0, pr: 4}}
            disabled={currentPath === "" || currentPath === "."}
            onClick={() => {
              setCurrentPath("");
              resetScroll();
            }}
          >
            <HomeIcon/>
          </IconButton>
          <Typography variant="h6" sx={{fontSize: "1rem", flexGrow: 1}}>
            {formatCurrentPath(currentPath)}
          </Typography>
          <SortFiles sort={sort} setSort={setSort}/>
          <CreateFolder
            path={currentPath}
            contents={listMenuFolder.data?.items}
            refresh={() => listMenuFolder.refetch()}
          />
        </Stack>
      </Paper>
      <Box sx={{height: "55px"}}></Box>
      {listMenuFolder.isLoading ? <LinearProgress/> : null}
      <Box>
        {!listMenuFolder.isLoading ? (
          <List>
            {listMenuFolder.data?.up ? (
              <ListItemButton
                onClick={() => {
                  setCurrentPath(
                    listMenuFolder.data.up ? listMenuFolder.data.up : ""
                  );
                  resetScroll();
                }}
              >
                <ListItemIcon>
                  <ArrowBackIcon/>
                </ListItemIcon>
                <ListItemText primary="Go back"/>
              </ListItemButton>
            ) : null}
            {sortedItems.map((item) => (
              <ListItem
                disablePadding
                key={item.path}
                secondaryAction={
                  <EditFile
                    item={item}
                    refresh={listMenuFolder.refetch}
                    parentContents={
                      listMenuFolder.data ? listMenuFolder.data.items : []
                    }
                  />
                }
              >
                <ListItemButton
                  onClick={() => {
                    if (item.next) {
                      setCurrentPath(item.next);
                      resetScroll();
                    } else {
                      api.launchFile(item.path).catch((err) => {
                        console.error(err);
                      });
                    }
                  }}
                >
                  <ListItemIcon>{icon(item)}</ListItemIcon>
                  <ListItemText
                    primary={item.namesTxt ? item.namesTxt : item.name}
                    secondary={
                      item.version
                        ? moment(item.version).format("YYYY-MM-DD")
                        : null
                    }
                  />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        ) : null}
        <ScrollToTopFab/>
      </Box>
    </>
  );
}

export function GamesMenu() {
  const api = new ControlApi();


  const [openShortcut, setOpenShortcut] = React.useState(false);
  const [selectedPath, setSelectedPath] = React.useState("");

  const [currentPath, setCurrentPath] = useState<string>("");
  const [sort, setSort] = useState<Sort>(Sort.NameAsc);
  const listGamesMenu = useListGamesFolder(currentPath);
  const [sortedItems, setSortedItems] = useState<MEMenuItem[]>(listGamesMenu.data?.items || [])

  useLayoutEffect(() => {
    setSortedItems(sortItems(listGamesMenu.data?.items, sort, currentPath));
  }, [sort, currentPath, listGamesMenu.data?.items])

  const icon = (item: MEMenuItem) => {
    switch (item.type) {
      case "folder":
        return <FolderOpenIcon/>;
      case "rbf":
        return <DeveloperBoardIcon/>;
      case "zip":
        return <FolderZipIcon/>;
      default:
        return <VideogameAssetIcon/>;
    }
  };

  const noEdit = currentPath === "" || currentPath === "." || listGamesMenu.data?.items.length === 0 || listGamesMenu.data?.items[0].inZip;

  const resetScroll = () => {
    window.scrollTo({
      top: 0,
    });
  };

  if (openShortcut && selectedPath !== "") {
    return (
      <SingleShortcut
        path={selectedPath}
        back={() => setOpenShortcut(false)}
      />
    );
  }

  return (
    <>
      <Paper
        sx={{
          boxShadow: 2,
          position: "fixed",
          width: 1,
          height: "55px",
          zIndex: 1,
          borderRadius: 0,
        }}
      >
        <Stack
          direction="row"
          sx={{p: 1, pl: 2, alignItems: "center", height: "55px"}}
        >
          <IconButton
            sx={{ml: -1, mr: 3}}
            disabled={currentPath === "" || currentPath === "."}
            onClick={() => {
              setCurrentPath("");
              resetScroll();
            }}
          >
            <HomeIcon/>
          </IconButton>
          <Typography variant="h6" sx={{fontSize: "1rem", flexGrow: 1}}>
            {formatCurrentPath(currentPath)}
          </Typography>
          {!noEdit ? (
            <SortFiles sort={sort} setSort={setSort}/>
          ) : null}
          {!noEdit ? (
            <CreateFolder
              path={currentPath}
              contents={listGamesMenu.data?.items}
              refresh={() => listGamesMenu.refetch()}
            />
          ) : null}

        </Stack>
      </Paper>
      <Box sx={{height: "55px"}}></Box>
      {listGamesMenu.isLoading ? <LinearProgress/> : null}
      <Box>
        {!listGamesMenu.isLoading ? (
          <List>
            {listGamesMenu.data?.up || listGamesMenu.data?.up === "" ? (
              <ListItemButton
                onClick={() => {
                  setCurrentPath(
                    listGamesMenu.data?.up ? listGamesMenu.data.up : ""
                  );
                  resetScroll();
                }}
              >
                <ListItemIcon>
                  <ArrowBackIcon/>
                </ListItemIcon>
                <ListItemText primary="Go back"/>
              </ListItemButton>
            ) : null}
            {sortedItems.map((item) => (
              <ListItem
                disablePadding
                key={item.path}
                secondaryAction={listGamesMenu.data?.up || listGamesMenu.data?.up === "" ? (
                  <EditFile
                    item={item}
                    refresh={listGamesMenu.refetch}
                    parentContents={
                      listGamesMenu.data ? listGamesMenu.data.items : []
                    }
                    setOpenShortcut={setOpenShortcut}
                    setSelectedPath={setSelectedPath}
                  />
                ) : null}
              >
                <ListItemButton
                  onClick={() => {
                    if (item.next) {
                      setCurrentPath(item.next);
                      resetScroll();
                    } else {
                      api.launchFile(item.path).catch((err) => {
                        console.error(err);
                      });
                    }
                  }}
                >
                  <ListItemIcon>{icon(item)}</ListItemIcon>
                  <ListItemText
                    primary={item.namesTxt ? item.namesTxt : item.name}
                    secondary={
                      item.version
                        ? moment(item.version).format("YYYY-MM-DD")
                        : null
                    }
                  />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        ) : null}
        <ScrollToTopFab/>
      </Box>
    </>
  );
}
