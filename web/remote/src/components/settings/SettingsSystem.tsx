import Stack from "@mui/material/Stack";
import {
  BoolOption,
  NumberOption,
  PageHeader,
  SaveButton,
  SimpleSelectOption,
  TextOption,
} from "./SettingsCommon";
import { useIniSettingsStore } from "../../lib/ini";
import TextField from "@mui/material/TextField";
import FormHelperText from "@mui/material/FormHelperText";
import FormControl from "@mui/material/FormControl";
import Button from "@mui/material/Button";
import { ControlApi } from "../../lib/api";
import React from "react";
import Dialog from "@mui/material/Dialog";
import DialogContent from "@mui/material/DialogContent";
import DialogContentText from "@mui/material/DialogContentText";
import DialogActions from "@mui/material/DialogActions";

function FbSize() {
  const v = useIniSettingsStore((state) => state.fbSize);
  const sv = useIniSettingsStore((state) => state.setFbSize);

  return (
    <SimpleSelectOption
      value={v}
      setValue={sv}
      options={[
        "Automatic",
        "Full size",
        "1/2 of resolution",
        "1/4 of resolution",
      ]}
      label={"Framebuffer size"}
      helpText="Set the final resolution of the Linux console displayed outside cores. A high framebuffer resolution can result in minor graphical glitches."
    />
  );
}

function FbTerminal() {
  const v = useIniSettingsStore((state) => state.fbTerminal);
  const sv = useIniSettingsStore((state) => state.setFbTerminal);

  return (
    <BoolOption
      value={v}
      setValue={sv}
      label={"Enable Linux framebuffer"}
      helpText="Linux console used for scripts, ARM ports and the menu wallpaper."
    />
  );
}

function BootCore() {
  const v = useIniSettingsStore((state) => state.bootCore);
  const sv = useIniSettingsStore((state) => state.setBootCore);
  // const [enabled, setEnabled] = useState(v && v !== "" ? true : false);
  //
  // // TODO: this style of value picker won't work because autoboot is a special option
  // //       maybe a base version that accepts options with an explicit label
  //
  // return (
  //   <FormControl>
  //     <FormControlLabel
  //       control={
  //         <Checkbox
  //           checked={enabled}
  //           onChange={(e) => {
  //             setEnabled(e.target.checked);
  //             if (!e.target.checked) {
  //               sv("");
  //             }
  //           }}
  //         />
  //       }
  //       label="Auto-boot core"
  //     />
  //     {enabled ? (
  //       <ValuePicker
  //         value={v}
  //         setValue={sv}
  //         options={[
  //           "Interpolation (Medium).txt",
  //           "No Interpolation.txt",
  //           "Scanlines (Medium).txt",
  //           "Scanlines (Strong).txt",
  //           "Scanlines (Weak).txt",
  //         ]}
  //         formatOption={(option) => option.replace(/\.[^/.]+$/, "")}
  //       />
  //     ) : null}
  //   </FormControl>
  // );

  return <TextOption value={v} setValue={sv} label="Auto-boot core" />;
}

function BootCoreTimeout() {
  const v = useIniSettingsStore((state) => state.bootCoreTimeout);
  const sv = useIniSettingsStore((state) => state.setBootCoreTimeout);
  const pv = useIniSettingsStore((state) => state.bootCore);

  return pv !== "" ? (
    <NumberOption
      value={v}
      setValue={sv}
      label="Auto-boot core timeout"
      defaultValue={"10"}
      min={10}
      max={30}
      disabledValue={"0"}
    />
  ) : null;
}

function WaitMount() {
  const v = useIniSettingsStore((state) => state.waitMount);
  const sv = useIniSettingsStore((state) => state.setWaitMount);

  // TODO: value picker of mount points
  return <TextOption value={v} setValue={sv} label="Wait for mount" />;
}

function Hostname() {
  const v = useIniSettingsStore((state) => state.hostname);
  const sv = useIniSettingsStore((state) => state.setHostname);

  return (
    <TextOption
      value={v}
      setValue={sv}
      label="Hostname"
      helpText="The name MiSTer will identify itself with on your network. MiSTer must be power cycled before this change takes effect."
    />
  );
}

function MACAddress() {
  const v = useIniSettingsStore((state) => state.ethernetMacAddress);
  const sv = useIniSettingsStore((state) => state.setEthernetMacAddress);
  const api = new ControlApi();

  return (
    <FormControl>
      <Stack direction="row">
        <TextField
          value={v}
          onChange={(e) => sv(e.target.value)}
          label="Ethernet MAC address"
          inputProps={{ maxLength: 17 }}
        />
        <Button
          sx={{ ml: 1 }}
          variant="outlined"
          onClick={() =>
            api
              .generateMac()
              .then((data) => sv(data.mac ? data.mac.toUpperCase() : ""))
          }
        >
          Generate
        </Button>
      </Stack>
      <FormHelperText>
        By default, all MiSTers have the same Ethernet MAC address, which will
        cause critical issues when there is more than one wired on the same
        network. MiSTer must be power cycled before this change takes effect.
      </FormHelperText>
    </FormControl>
  );
}

function Reboot() {
  const api = new ControlApi();
  const [open, setOpen] = React.useState(false);
  const handleOpen = () => setOpen(true);
  const handleClose = () => setOpen(false);

  return (
    <>
      <Button variant="outlined" onClick={handleOpen}>
        Reboot MiSTer
      </Button>
      <Dialog open={open} onClose={handleClose}>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to reboot MiSTer? This will immediately quit
            any running cores.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleClose}>Cancel</Button>
          <Button
            onClick={() => {
              api.reboot().then((r) => handleClose());
            }}
            autoFocus
          >
            Reboot
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
}

export default function SystemSettings() {
  return (
    <>
      <PageHeader title="System" />
      <Stack spacing={3} m={2}>
        <Hostname />
        <MACAddress />
        <FbTerminal />
        <FbSize />
        <BootCore />
        <BootCoreTimeout />
        <WaitMount />
        <Reboot />
      </Stack>
      <SaveButton />
    </>
  );
}
