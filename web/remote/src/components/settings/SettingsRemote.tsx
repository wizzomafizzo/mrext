import FormControl from "@mui/material/FormControl";
import InputLabel from "@mui/material/InputLabel";
import MenuItem from "@mui/material/MenuItem";
import Select from "@mui/material/Select";
import Stack from "@mui/material/Stack";
import { useUIStateStore } from "../../lib/store";
import { themes } from "../../lib/themes";
import { PageHeader, SectionHeader } from "./SettingsCommon";
import TextField from "@mui/material/TextField";
import Button from "@mui/material/Button";
import React, { useState } from "react";
import {
  ControlApi,
  getApiEndpoint,
  getStoredApiEndpoint,
  setApiEndpoint,
  getStoredWsEndpoint,
  getWsEndpoint,
  setWsEndpoint,
} from "../../lib/api";
import { Card } from "@mui/material";
import FormLabel from "@mui/material/FormLabel";
import RemoveIcon from "@mui/icons-material/Remove";
import AddIcon from "@mui/icons-material/Add";
import { Add } from "@mui/icons-material";
import Typography from "@mui/material/Typography";
import FormHelperText from "@mui/material/FormHelperText";

export default function Remote() {
  const activeTheme = useUIStateStore((state) => state.activeTheme);
  const setActiveTheme = useUIStateStore((state) => state.setActiveTheme);
  const fontSize = useUIStateStore((state) => state.fontSize);
  const setFontSize = useUIStateStore((state) => state.setFontSize);

  const api = new ControlApi();

  const storedApiEndpoint = getStoredApiEndpoint();
  const [settingApiEndpoint, setSettingApiEndpoint] = useState(
    storedApiEndpoint ? storedApiEndpoint : ""
  );

  const storedWsEndpoint = getStoredWsEndpoint();
  const [settingWsEndpoint, setSettingWsEndpoint] = useState(
    storedWsEndpoint ? storedWsEndpoint : ""
  );

  let storedAddress = "";
  if (storedApiEndpoint && storedApiEndpoint !== "") {
    const url = new URL(storedApiEndpoint);
    storedAddress = url.hostname;
  }

  const [address, setAddress] = useState(storedAddress);
  const saveAddress = (address: string) => {
    const url = "http://" + address + ":8182/api";
    setApiEndpoint(url);
    location.reload();
  };

  return (
    <>
      <PageHeader title="Remote" noRevert />
      <Stack spacing={3} m={2}>
        <FormControl>
          <InputLabel>Theme</InputLabel>
          <Select
            value={activeTheme}
            label="Theme"
            onChange={(e) => setActiveTheme(e.target.value)}
          >
            {Object.keys(themes).map((id) => (
              <MenuItem key={id} value={id}>
                {themes[id].displayName}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <FormControl>
          <FormLabel>Font size</FormLabel>
          <Stack direction="row" spacing={1}>
            <Button
              variant="outlined"
              onClick={() => {
                if (fontSize > 5) {
                  setFontSize(fontSize - 1);
                }
              }}
            >
              <RemoveIcon />
            </Button>
            <Typography sx={{ pl: 2, pr: 2, fontSize: "150%" }}>
              {fontSize}
            </Typography>
            <Button
              variant="outlined"
              onClick={() => {
                if (fontSize < 50) {
                  setFontSize(fontSize + 1);
                }
              }}
            >
              <AddIcon />
            </Button>
            <Button
              variant="outlined"
              onClick={() => {
                setFontSize(14);
              }}
            >
              Reset
            </Button>
          </Stack>
        </FormControl>

        <FormControl sx={{ pt: 1.5 }}>
          <Stack spacing={1}>
            <TextField
              label="MiSTer address"
              value={address}
              onChange={(event) => setAddress(event.target.value)}
              fullWidth
            />
            <FormHelperText>
              The IP address or hostname of the connected MiSTer. Use the API
              endpoint settings in the advanced section below for finer control.
            </FormHelperText>
            <Button
              fullWidth
              variant="outlined"
              onClick={() => saveAddress(address)}
            >
              Save address
            </Button>
          </Stack>
        </FormControl>

        <SectionHeader text={"Advanced"} />

        <FormControl>
          <Button
            variant="outlined"
            href={getApiEndpoint() + "/settings/remote/log"}
            target="_blank"
          >
            Download log file
          </Button>
        </FormControl>

        <FormControl>
          <Button
            variant="outlined"
            onClick={() => {
              api.restartRemoteService().catch((e) => {
                console.error(e);
              });
            }}
          >
            Restart remote service
          </Button>
        </FormControl>

        <Card>
          <Stack m={1} spacing={1}>
            <FormControl>
              <TextField
                label="API endpoint URL"
                value={settingApiEndpoint}
                placeholder={getApiEndpoint()}
                onChange={(e) => setSettingApiEndpoint(e.target.value)}
                autoComplete="off"
              />
            </FormControl>
            <FormControl>
              <TextField
                label="WebSocket endpoint URL"
                value={settingWsEndpoint}
                placeholder={getWsEndpoint()}
                onChange={(e) => setSettingWsEndpoint(e.target.value)}
                autoComplete="off"
                helperText="Leave blank to auto-configure based on API endpoint URL."
              />
            </FormControl>
            <Button
              variant="outlined"
              onClick={() => {
                setApiEndpoint(settingApiEndpoint);
                setWsEndpoint(settingWsEndpoint);
                window.location.reload();
              }}
            >
              Save endpoints
            </Button>
          </Stack>
        </Card>
      </Stack>
    </>
  );
}
