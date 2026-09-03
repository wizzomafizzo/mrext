import List from "@mui/material/List";
import ListItem from "@mui/material/ListItem";
import ListItemButton from "@mui/material/ListItemButton";
import ListItemText from "@mui/material/ListItemText";
import ListItemIcon from "@mui/material/ListItemIcon";
import ArrowForwardIcon from "@mui/icons-material/ArrowForward";
import DeveloperBoardIcon from "@mui/icons-material/DeveloperBoard";
import TvIcon from "@mui/icons-material/Tv";
import WysiwygIcon from "@mui/icons-material/Wysiwyg";
import SpeakerIcon from "@mui/icons-material/Speaker";
import SportsEsportsIcon from "@mui/icons-material/SportsEsports";
import SettingsRemoteIcon from "@mui/icons-material/SettingsRemote";
import FilterIcon from "@mui/icons-material/Filter";
import SettingsInputCompositeIcon from "@mui/icons-material/SettingsInputComposite";
import ComputerIcon from "@mui/icons-material/Computer";

import { SettingsPageId, useUIStateStore } from "../../lib/store";

import {
  GeneralVideoSettings,
  VideoFiltersSettings,
  AnalogVideoSettings,
} from "./SettingsVideo";
import InputDevices from "./SettingsInputDevices";
import OSDMenuSettings from "./SettingsOSDMenu";
import CoresSettings from "./SettingsCore";
import Remote from "./SettingsRemote";

import Box from "@mui/material/Box";
import AudioSettings from "./SettingsAudio";
import SystemSettings from "./SettingsSystem";
import { ReactNode, useEffect, useState } from "react";
import { useListMisterInis } from "../../lib/queries";
import Button from "@mui/material/Button";
import SwapHorizIcon from "@mui/icons-material/SwapHoriz";
import Dialog from "@mui/material/Dialog";
import Typography from "@mui/material/Typography";
import { ControlApi } from "../../lib/api";
import { loadMisterIni, useIniSettingsStore } from "../../lib/ini";

function SettingsPageLink(props: {
  page: SettingsPageId;
  text: string;
  icon: ReactNode;
}) {
  const setActiveSettingsPage = useUIStateStore(
    (state) => state.setActiveSettingsPage
  );

  return (
    <ListItem disableGutters sx={{ p: 0, pt: 1 }}>
      <ListItemButton onClick={() => setActiveSettingsPage(props.page)}>
        <ListItemIcon>{props.icon}</ListItemIcon>
        <ListItemText primary={props.text} />
        <ArrowForwardIcon />
      </ListItemButton>
    </ListItem>
  );
}

function IniSwitcher() {
  const api = new ControlApi();
  const inis = useListMisterInis();
  const [iniDialogOpen, setIniDialogOpen] = useState(false);
  const iniSettings = useIniSettingsStore();

  let activeIni = {
    name: "Main",
    id: 1,
  };

  if (inis.data && inis.data.inis && inis.data.inis.length > 0) {
    let id: number;

    if (inis.data.active === 0) {
      id = 1;
    } else {
      id = inis.data.active;
    }

    activeIni.name = inis.data.inis[id - 1].displayName;
    activeIni.id = id;
  }

  return (
    <>
      <Dialog open={iniDialogOpen} onClose={() => setIniDialogOpen(false)}>
        <Box sx={{ p: 2, minWidth: 200 }}>
          <List disablePadding>
            <ListItem sx={{ pt: 0 }}>
              <Typography
                sx={{
                  fontWeight: 500,
                  textAlign: "center",
                  width: 1,
                }}
              >
                Set active INI
              </Typography>
            </ListItem>
            {(inis.data?.inis || []).map(
              (
                ini: {
                  filename: string;
                  displayName: string;
                },
                i: number
              ) => (
                <ListItem key={ini.filename} disableGutters>
                  <Button
                    fullWidth
                    variant={activeIni.id == i + 1 ? "contained" : "outlined"}
                    onClick={() => {
                      setIniDialogOpen(false);
                      api.setMisterIni({ ini: i + 1 }).then(() => {
                        inis.refetch().catch((e) => console.error(e));
                        loadMisterIni(i + 1, iniSettings, true).catch((e) =>
                          console.error(e)
                        );
                      });
                    }}
                  >
                    {ini.displayName}
                  </Button>
                </ListItem>
              )
            )}
          </List>
        </Box>
      </Dialog>
      <ListItem sx={{ pb: 0, pt: 2 }}>
        <Button
          fullWidth
          variant="contained"
          endIcon={<SwapHorizIcon />}
          onClick={() => setIniDialogOpen(true)}
          disabled={!inis.data || !inis.data.inis || inis.data.inis.length < 2}
        >
          Active INI: {activeIni.name}
        </Button>
      </ListItem>
    </>
  );
}

function MainPage() {
  return (
    <div>
      <List disablePadding>
        <IniSwitcher />
        <SettingsPageLink
          page={SettingsPageId.GeneralVideo}
          text="General Video"
          icon={<TvIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.VideoFilters}
          text="Video Filters"
          icon={<FilterIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.AnalogVideo}
          text="Analog Video"
          icon={<SettingsInputCompositeIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.Audio}
          text="Audio"
          icon={<SpeakerIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.InputDevices}
          text="Input Devices"
          icon={<SportsEsportsIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.OSDMenu}
          text="OSD and Menu"
          icon={<WysiwygIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.Cores}
          text="Cores"
          icon={<DeveloperBoardIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.System}
          text="System"
          icon={<ComputerIcon />}
        />
        <SettingsPageLink
          page={SettingsPageId.Remote}
          text="Remote"
          icon={<SettingsRemoteIcon />}
        />
      </List>
    </div>
  );
}

export default function Settings() {
  const activeSettingsPage = useUIStateStore(
    (state) => state.activeSettingsPage
  );

  const iniSettingsStore = useIniSettingsStore();

  useEffect(() => {
    const api = new ControlApi();
    api.listMisterInis().then((inis) => {
      let id: number;
      if (inis.active === 0) {
        id = 1;
      } else {
        id = inis.active;
      }

      loadMisterIni(id, iniSettingsStore).catch((err) => console.error(err));
    });
  }, []);

  const page = (() => {
    switch (activeSettingsPage) {
      case SettingsPageId.Main:
        return <MainPage />;
      case SettingsPageId.GeneralVideo:
        return <GeneralVideoSettings />;
      case SettingsPageId.Audio:
        return <AudioSettings />;
      case SettingsPageId.InputDevices:
        return <InputDevices />;
      case SettingsPageId.OSDMenu:
        return <OSDMenuSettings />;
      case SettingsPageId.Cores:
        return <CoresSettings />;
      case SettingsPageId.System:
        return <SystemSettings />;
      case SettingsPageId.Remote:
        return <Remote />;
      case SettingsPageId.VideoFilters:
        return <VideoFiltersSettings />;
      case SettingsPageId.AnalogVideo:
        return <AnalogVideoSettings />;
    }
  })();

  return <Box>{page}</Box>;
}
