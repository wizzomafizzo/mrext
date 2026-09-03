import React, { ReactNode, useEffect, useState } from "react";

import AppBar from "@mui/material/AppBar";
import Box from "@mui/material/Box";
import CssBaseline from "@mui/material/CssBaseline";
import Divider from "@mui/material/Divider";
import Drawer from "@mui/material/Drawer";
import IconButton from "@mui/material/IconButton";
import List from "@mui/material/List";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemIcon from "@mui/material/ListItemIcon";
import ListItemText from "@mui/material/ListItemText";
import MenuIcon from "@mui/icons-material/Menu";
import Toolbar from "@mui/material/Toolbar";
import Typography from "@mui/material/Typography";
import { useMediaQuery } from "react-responsive";

import DashboardIcon from "@mui/icons-material/Dashboard";
import SearchIcon from "@mui/icons-material/Search";
import VideogameAssetIcon from "@mui/icons-material/VideogameAsset";
import GamepadIcon from "@mui/icons-material/Gamepad";
import SettingsIcon from "@mui/icons-material/Settings";
import PhotoCameraBackIcon from "@mui/icons-material/PhotoCameraBack";
import FormatPaintIcon from "@mui/icons-material/FormatPaint";
import MusicNoteIcon from "@mui/icons-material/MusicNote";
import DvrIcon from "@mui/icons-material/Dvr";
import SignalWifiBadIcon from "@mui/icons-material/SignalWifiBad";
import PlayArrowIcon from "@mui/icons-material/PlayArrow";
import PauseIcon from "@mui/icons-material/Pause";
import LanIcon from "@mui/icons-material/Lan";
import TerminalIcon from "@mui/icons-material/Terminal";
import DeveloperBoardIcon from "@mui/icons-material/DeveloperBoard";

import {
  Navigate,
  NavLink,
  NavLinkProps,
  Route,
  Routes,
  useLocation,
  useNavigate,
} from "react-router-dom";

import Screenshots from "./Screenshots";
import Systems from "./Systems";
import Wallpaper from "./Wallpaper";
import Music from "./Music";
import Search from "./Search";
import Settings from "./settings/Settings";
import Control from "./Control";
import {
  ClickAwayListener,
  Grow,
  Paper,
  Popper,
  SwipeableDrawer,
} from "@mui/material";
import {GamesMenu, Menu} from "./Menu";
import useWs from "./WebSocket";
import {
  SettingsPageId,
  useServerStateStore,
  useUIStateStore,
} from "../lib/store";
import Stack from "@mui/material/Stack";
import Button from "@mui/material/Button";
import {
  ControlApi,
  getApiEndpoint,
  getStoredApiEndpoint,
  setApiEndpoint,
} from "../lib/api";
import { useQuery } from "@tanstack/react-query";
import ListItem from "@mui/material/ListItem";
import moment from "moment";
import { Network } from "./Network";
import { Scripts } from "./Scripts";
import { useMusicStatus } from "../lib/queries";
import LinearProgress from "@mui/material/LinearProgress";
import { ControlAuto } from "./ControlAuto";
import Grid from "@mui/material/Grid";
import Dialog from "@mui/material/Dialog";
import TextField from "@mui/material/TextField";
import { Capacitor } from "@capacitor/core";

const drawerWidth = 240;

interface Page {
  path: string;
  titleText: string;
  buttonText: string;
  icon: ReactNode;
}

const pages: Page[] = [
  {
    path: "/",
    titleText: "Dashboard",
    buttonText: "Dashboard",
    icon: <DashboardIcon />,
  },
  {
    path: "/search",
    titleText: "Search",
    buttonText: "Search",
    icon: <SearchIcon />,
  },
  {
    path: "/games",
    titleText: "Games",
    buttonText: "Games",
    icon: <VideogameAssetIcon />,
  },
  {
    path: "/systems",
    titleText: "Systems",
    buttonText: "Systems",
    icon: <DeveloperBoardIcon />,
  },
  {
    path: "/screenshots",
    titleText: "Screenshots",
    buttonText: "Screenshots",
    icon: <PhotoCameraBackIcon />,
  },
  {
    path: "/control",
    titleText: "Control",
    buttonText: "Control",
    icon: <GamepadIcon />,
  },
  {
    path: "/control/auto",
    titleText: "Gamepad",
    buttonText: "Gamepad",
    icon: <GamepadIcon />,
  },
  {
    path: "/settings",
    titleText: "Settings",
    buttonText: "Settings",
    icon: <SettingsIcon />,
  },
  {
    path: "/wallpaper",
    titleText: "Wallpaper",
    buttonText: "Wallpaper",
    icon: <FormatPaintIcon />,
  },
  {
    path: "/music",
    titleText: "Music",
    buttonText: "Music",
    icon: <MusicNoteIcon />,
  },
  {
    path: "/menu",
    titleText: "Menu",
    buttonText: "Menu",
    icon: <DvrIcon />,
  },
  {
    path: "/network",
    titleText: "Network",
    buttonText: "Network",
    icon: <LanIcon />,
  },
  {
    path: "/scripts",
    titleText: "Scripts",
    buttonText: "Scripts",
    icon: <TerminalIcon />,
  },
];

function getPage(path: string): Page | undefined {
  for (const page of pages) {
    if (page.path === path) {
      return page;
    }
  }
}

type RouterLinkProps = React.PropsWithChildren<{
  to: string;
  text: string;
  icon: ReactNode;
  onClick?: () => void;
  closeDrawer: () => void;
}>;

function RouterLink(props: RouterLinkProps) {
  type MyNavLinkProps = Omit<NavLinkProps, "to">;
  const MyNavLink = React.useMemo(
    () =>
      React.forwardRef<HTMLAnchorElement, MyNavLinkProps>(
        (navLinkProps, ref) => {
          const { className: previousClasses, ...rest } = navLinkProps;
          const elementClasses = previousClasses?.toString() ?? "";
          return (
            <NavLink
              {...rest}
              ref={ref}
              to={props.to}
              end
              className={({ isActive }) =>
                isActive ? elementClasses + " Mui-selected" : elementClasses
              }
            />
          );
        }
      ),
    [props.to]
  );
  return (
    <ListItemButton
      component={MyNavLink}
      onClick={() => {
        if (props.onClick) {
          props.onClick();
        }
        props.closeDrawer();
      }}
    >
      <ListItemIcon
        sx={{
          ".Mui-selected > &": {
            color: (theme) => theme.palette.primary.main,
          },
        }}
      >
        {props.icon}
      </ListItemIcon>
      <ListItemText primary={props.text} />
    </ListItemButton>
  );
}

function PlayingButton(props: {
  connectOpen: boolean;
  setConnectOpen: (open: boolean) => void;
}) {
  const ws = useWs();
  const serverState = useServerStateStore();
  let icon = <PauseIcon />;
  const anchorRef = React.useRef<HTMLButtonElement>(null);
  const [open, setOpen] = React.useState(false);
  const api = new ControlApi();

  const handleClose = (event: Event) => {
    if (
      anchorRef.current &&
      anchorRef.current.contains(event.target as HTMLElement)
    ) {
      return;
    }

    setOpen(false);
  };

  if (ws.readyState !== WebSocket.OPEN) {
    icon = <SignalWifiBadIcon />;
  } else if (serverState.activeGame !== "" || serverState.activeCore !== "") {
    icon = <PlayArrowIcon />;
  } else {
    icon = <PauseIcon />;
  }

  const parseActiveGame = (activeGame: string) => {
    const game = activeGame.split("/").pop();
    return game?.split(".").shift();
  };

  return (
    <>
      <Button
        ref={anchorRef}
        sx={{
          color: "primary.contrastText",
          minWidth: "40px",
          pl: "5px",
          pr: "5px",
          mr: 0,
        }}
        onClick={() => {
          if (ws.readyState !== WebSocket.OPEN) {
            props.setConnectOpen(true);
          } else {
            setOpen(!open);
          }
        }}
        variant="outlined"
        size="small"
      >
        {ws.readyState !== WebSocket.OPEN && (
          <span style={{ paddingRight: "5px", paddingTop: "2px" }}>
            Connect
          </span>
        )}{" "}
        {icon}
      </Button>
      <Popper
        sx={{
          zIndex: 1,
        }}
        open={open}
        anchorEl={anchorRef.current}
        role={undefined}
        transition
        disablePortal
      >
        {({ TransitionProps, placement }) => (
          <Grow
            {...TransitionProps}
            style={{
              transformOrigin:
                placement === "bottom" ? "center top" : "center bottom",
            }}
          >
            <Paper sx={{ mr: 0 }}>
              <ClickAwayListener onClickAway={handleClose}>
                <Stack sx={{ p: 1 }}>
                  <Typography>
                    Core: {serverState.activeCore || "None"}
                  </Typography>
                  <Typography>
                    Game: {parseActiveGame(serverState.activeGame) || "None"}
                  </Typography>
                  {serverState.activeCore ? (
                    <Button
                      onClick={() => {
                        api.launchMenu();
                        setOpen(false);
                      }}
                    >
                      Exit to menu
                    </Button>
                  ) : null}
                </Stack>
              </ClickAwayListener>
            </Paper>
          </Grow>
        )}
      </Popper>
    </>
  );
}

function humanFileSize(bytes: number) {
  if (bytes === 0) {
    return "0 B";
  }

  const thresh = 1024;
  let unit = 0;

  while (bytes >= thresh || -bytes >= thresh) {
    bytes /= thresh;
    unit++;
  }

  return (unit ? bytes.toFixed(1) + "" : bytes) + " KMGTPEZY"[unit] + "B";
}

function ConnectDialog(props: {
  open: boolean;
  setOpen: (open: boolean) => void;
}) {
  const apiUrl = getStoredApiEndpoint();
  let storedAddress = "";
  if (apiUrl && apiUrl !== "") {
    const url = new URL(apiUrl);
    storedAddress = url.hostname;
  }

  const [address, setAddress] = useState(storedAddress);
  const [canConnect, setCanConnect] = useState(false);

  const tryConnect = (address: string) => {
    const url = "http://" + address + ":8182/api";
    const statusUrl = url + "/sysinfo";
    fetch(statusUrl).then((response) => {
      if (response.status === 200) {
        setCanConnect(true);
      }
    });
  };

  const saveAddress = (address: string) => {
    const url = "http://" + address + ":8182/api";
    setApiEndpoint(url);
    props.setOpen(false);
    location.reload();
  };

  useEffect(() => {
    setCanConnect(false);
    tryConnect(address);
  }, [address]);

  return (
    <Dialog open={props.open} onClose={() => props.setOpen(false)}>
      <Stack sx={{ p: 1 }} spacing={2}>
        <Typography variant="body2">
          Enter your MiSTer's IP address or hostname. The{" "}
          <a
            style={{ color: "white" }}
            target="_blank"
            href="https://github.com/wizzomafizzo/mrext/blob/main/docs/remote.md"
          >
            Remote script
          </a>{" "}
          must already be installed and running on the MiSTer.
        </Typography>
        <TextField
          label="MiSTer address"
          value={address}
          onChange={(event) => setAddress(event.target.value)}
          fullWidth
        />
        <Typography variant="body2">
          This can be changed in the <i>Remote</i> section of the{" "}
          <i>Settings</i> page.
        </Typography>
        <Button
          disabled={!canConnect}
          fullWidth
          variant="outlined"
          color="success"
          onClick={() => saveAddress(address)}
        >
          Save
        </Button>
      </Stack>
    </Dialog>
  );
}

export default function ResponsiveDrawer() {
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = React.useState(false);
  const navigate = useNavigate();
  const api = new ControlApi();
  const ws = useWs();
  const uiState = useUIStateStore();
  const sysInfo = useQuery({
    queryKey: ["sysInfo"],
    queryFn: api.sysInfo,
  });
  const peers = useQuery({
    queryKey: ["settings", "remote", "peers"],
    queryFn: api.getPeers,
  });
  const musicStatus = useMusicStatus(true);
  const isMobile = useMediaQuery({ query: "(max-width: 600px)" });
  const [connectOpen, setConnectOpen] = useState(false);

  useEffect(() => {
    if (Capacitor.isNativePlatform()) {
      const storedApiUrl = getStoredApiEndpoint();
      if (!storedApiUrl || storedApiUrl === "") {
        setConnectOpen(true);
      }
    }
  }, []);

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const drawer = (
    <div>
      <Toolbar sx={{ justifyContent: "center", paddingTop: "env(safe-area-inset-top)" }}>
        <Stack>
          <Stack direction="row">
            <img alt="MiSTer Kun Logo" src="/misterkun.svg" height={43} />
            <img
              alt="MiSTer FPGA Logo"
              src={getApiEndpoint() + "/settings/remote/logo"}
              height={43}
            />
          </Stack>
          {sysInfo.data ? (
            <Typography fontSize="x-small" textAlign="center">
              {sysInfo.data.hostname} v{sysInfo.data.version}
            </Typography>
          ) : null}
        </Stack>
      </Toolbar>
      <List>
        {/*<RouterLink to="/" text="Dashboard" icon={<DashboardIcon />} closeDrawer={handleDrawerToggle} />*/}
        <RouterLink
          to="/control"
          text="Control"
          icon={<GamepadIcon />}
          closeDrawer={handleDrawerToggle}
        />
        <RouterLink
          to="menu"
          text="Menu"
          icon={<DvrIcon />}
          closeDrawer={handleDrawerToggle}
        />
      </List>
      <Divider />
      <List>
        <RouterLink
          to="/search"
          text="Search"
          icon={<SearchIcon />}
          closeDrawer={handleDrawerToggle}
        />
        <RouterLink to="/games" text="Games" icon={<VideogameAssetIcon />} closeDrawer={handleDrawerToggle} />
        <RouterLink
          to="/systems"
          text="Systems"
          icon={<DeveloperBoardIcon />}
          closeDrawer={handleDrawerToggle}
        />
        <RouterLink
          to="/screenshots"
          text="Screenshots"
          icon={<PhotoCameraBackIcon />}
          closeDrawer={handleDrawerToggle}
        />
      </List>
      <Divider />
      <List>
        {musicStatus.data?.running && (
          <RouterLink
            to="/music"
            text="Music"
            icon={<MusicNoteIcon />}
            closeDrawer={handleDrawerToggle}
          />
        )}
        <RouterLink
          to="/wallpaper"
          text="Wallpaper"
          icon={<FormatPaintIcon />}
          closeDrawer={handleDrawerToggle}
        />
        <RouterLink
          to="/scripts"
          text="Scripts"
          icon={<TerminalIcon />}
          closeDrawer={handleDrawerToggle}
        />
        <RouterLink
          to="/settings"
          text="Settings"
          icon={<SettingsIcon />}
          onClick={() => uiState.setActiveSettingsPage(SettingsPageId.Main)}
          closeDrawer={handleDrawerToggle}
        />
      </List>
      <Divider />
      {sysInfo.data && sysInfo.data.ips && sysInfo.data.disks ? (
        <div onClick={() => sysInfo.refetch()}>
          <List dense sx={{ pb: 0.6 }}>
            {sysInfo.data.disks.map((disk) => (
              <Box key={disk.path}>
                <ListItem>
                  <ListItemText
                    primary={disk.displayName}
                    secondary={
                      humanFileSize(disk.free) +
                      " / " +
                      humanFileSize(disk.total) +
                      " available"
                    }
                  />
                </ListItem>
                <LinearProgress
                  variant="determinate"
                  value={Math.round((disk.used / disk.total) * 100)}
                  sx={{ ml: 2, mr: 2, mt: -1, mb: 1.5 }}
                />
              </Box>
            ))}
            <ListItem>
              <ListItemText
                primary="Last system update"
                secondary={
                  sysInfo.data.updated === ""
                    ? "Never"
                    : moment(sysInfo.data.updated).fromNow()
                }
              />
            </ListItem>
            <ListItem>
              <ListItemText
                primary={
                  sysInfo.data.ips.length > 1 ? "IP addresses" : "IP address"
                }
                secondary={sysInfo.data.ips.join(", ")}
              />
            </ListItem>
            <ListItem>
              <ListItemText
                primary="Hostname"
                secondary={
                  sysInfo.data.hostname +
                  (sysInfo.data.dns !== "" ? " (" + sysInfo.data.dns + ")" : "")
                }
              />
            </ListItem>
          </List>
        </div>
      ) : null}
      {peers.data && peers.data.peers && peers.data.peers.length > 1 ? (
        <Box sx={{ p: 1, pt: 0 }}>
          <Button
            sx={{ width: 1 }}
            variant="outlined"
            startIcon={<LanIcon />}
            onClick={() => {
              navigate("/network");
              setMobileOpen(false);
            }}
          >
            Browse MiSTers
          </Button>
        </Box>
      ) : null}
    </div>
  );

  // for some reason the min-height of a Toolbar component is not right on iOS and
  // is different between mobile and desktop (assuming it's actually between the hidden
  // and always visible sidebar layouts). this may also have something to to with safe
  // area insets on these devices. this is a hack to fix it. it causes the second floating
  // toolbar on menu and games pages to be offset slightly. it's probably best in the future
  // to come up with an alternative layout that doesn't use double fixed toolbars
  const toolbarStyle: {minHeight?: string} = {}
  if (Capacitor.getPlatform() === "ios" && isMobile) {
    toolbarStyle["minHeight"] = "50px";
  } else if (Capacitor.getPlatform() === "ios" && !isMobile) {
    toolbarStyle["minHeight"] = "40px";
  }

  return (
    <Box sx={{ display: "flex" }}>
      <CssBaseline />
      <AppBar
        position="fixed"
        sx={{
          width: isMobile ? "100%" : `calc(100% - ${drawerWidth}px)`,
          ml: isMobile ? "0" : `${drawerWidth}px`,
          // paddingTop: "env(safe-area-inset-top)",
        }}
      >
        <Toolbar
          sx={{
            backgroundColor: "secondary.main",
            color: "primary.contrastText",
            pl: 1,
            pr: 1,
            paddingTop: "env(safe-area-inset-top)",
          }}
        >
          <Grid
            container
            direction="row"
            justifyContent="flex-end"
            alignItems="center"
          >
            <Grid item xs={4}>
              <IconButton
                color="inherit"
                aria-label="open drawer"
                edge="start"
                onClick={handleDrawerToggle}
                sx={{
                  pt: "12px",
                  pl: 2,
                  display: isMobile ? "block" : "none",
                }}
              >
                <MenuIcon />
              </IconButton>
            </Grid>
            <Grid item xs={4} textAlign="center">
              <Typography variant="h6" noWrap component="div">
                {getPage(location.pathname)?.titleText}
              </Typography>
            </Grid>
            <Grid item xs={4} textAlign="right">
              <PlayingButton
                connectOpen={connectOpen}
                setConnectOpen={setConnectOpen}
              />
            </Grid>
          </Grid>
          <ConnectDialog open={connectOpen} setOpen={setConnectOpen} />
        </Toolbar>
      </AppBar>
      <Box
        component="nav"
        sx={!isMobile ? { width: drawerWidth, flexShrink: 0 } : null}
        aria-label="mailbox folders"
      >
        {isMobile ? (
          <SwipeableDrawer
            container={window.document.body}
            variant="temporary"
            open={mobileOpen}
            onClose={handleDrawerToggle}
            onOpen={handleDrawerToggle}
            disableBackdropTransition
            ModalProps={{
              keepMounted: true, // Better open performance on mobile.
            }}
            sx={{
              display: "block",
              "& .MuiDrawer-paper": {
                boxSizing: "border-box",
                width: drawerWidth,
              },
            }}
          >
            {drawer}
          </SwipeableDrawer>
        ) : (
          <Drawer
            variant="permanent"
            sx={{
              display: "block",
              "& .MuiDrawer-paper": {
                boxSizing: "border-box",
                width: drawerWidth,
              },
            }}
            open
          >
            {drawer}
          </Drawer>
        )}
      </Box>
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          width: isMobile ? "100%" : `calc(100% - ${drawerWidth}px)`,
        }}
      >
        <Toolbar style={toolbarStyle} />
        <Routes>
          <Route path="/systems" element={<Systems />} />
          <Route path="/" element={<Navigate to="/control" />} />
          <Route path="/search" element={<Search />} />
          <Route path="/screenshots" element={<Screenshots />} />
          <Route path="/control" element={<Control />} />
          <Route path="/music" element={<Music />} />
          <Route path="/wallpaper" element={<Wallpaper />} />
          <Route path="/settings" element={<Settings />} />
          <Route path="/menu" element={<Menu />} />
          <Route path="/network" element={<Network />} />
          <Route path="/scripts" element={<Scripts />} />
          <Route path="/control/auto" element={<ControlAuto />} />
          <Route path="/games" element={<GamesMenu />} />
        </Routes>
      </Box>
    </Box>
  );
}
