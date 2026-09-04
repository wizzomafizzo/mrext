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

function LogFileEntry() {
  const v = useIniSettingsStore((state) => state.logFileEntry);
  const sv = useIniSettingsStore((state) => state.setLogFileEntry);

  return (
    <BoolOption
      label="Log current file entry"
      value={v}
      setValue={sv}
      helpText="Log the currently selected item in menu to a file. Useful for scripts."
    />
  );
}

function RbfHideDatecode() {
  const v = useIniSettingsStore((state) => state.rbfHideDatecode);
  const sv = useIniSettingsStore((state) => state.setRbfHideDatecode);

  return (
    <BoolOption
      value={v}
      setValue={sv}
      label="Show core release dates"
      helpText="Display core's release date next to its menu entry. Press F2 to toggle temporarily."
      invert
    />
  );
}

function OsdRotate() {
  const v = useIniSettingsStore((state) => state.osdRotate);
  const sv = useIniSettingsStore((state) => state.setOsdRotate);

  return (
    <SimpleSelectOption
      value={v}
      setValue={sv}
      options={["No rotation", "Rotate right (+90°)", "Rotate left (-90°)"]}
      label="OSD rotation"
    />
  );
}

function BrowseExpand() {
  const v = useIniSettingsStore((state) => state.browseExpand);
  const sv = useIniSettingsStore((state) => state.setBrowseExpand);

  return (
    <BoolOption
      value={v}
      setValue={sv}
      label="Expand file browser selection"
      helpText="Long file names selected in the OSD and menu will be shown using two lines."
    />
  );
}

function OSDTimeout() {
  const v = useIniSettingsStore((state) => state.osdTimeout);
  const sv = useIniSettingsStore((state) => state.setOsdTimeout);

  return (
    <NumberOption
      value={v}
      setValue={sv}
      label="OSD display timeout"
      defaultValue={"30"}
      disabledValue={"0"}
      min={1}
      max={3600}
      helpText="Delay before OSD in the menu is hidden and screen is dimmed after inactivity."
      suffix="seconds"
    />
  );
}

function VideoOff() {
  const v = useIniSettingsStore((state) => state.videoOff);
  const sv = useIniSettingsStore((state) => state.setVideoOff);
  const pv = useIniSettingsStore((state) => state.osdTimeout);

  return Number(pv) > 0 ? (
    <NumberOption
      value={v}
      setValue={sv}
      label="Blank screen after timeout"
      defaultValue={"30"}
      disabledValue={"0"}
      min={1}
      max={3600}
      helpText="Delay before displaying a black screen after OSD display timeout."
      suffix="seconds"
    />
  ) : null;
}

function MenuPal() {
  const v = useIniSettingsStore((state) => state.menuPal);
  const sv = useIniSettingsStore((state) => state.setMenuPal);

  return (
    <BoolOption
      value={v}
      setValue={sv}
      label="PAL mode for menu"
      helpText="Default is NTSC mode."
    />
  );
}

function Font() {
  const v = useIniSettingsStore((state) => state.font);
  const sv = useIniSettingsStore((state) => state.setFont);
  // const [enabled, setEnabled] = useState(v !== "");
  //
  // return (
  //   <FormControl>
  //     <FormLabel>Menu font</FormLabel>
  //     <ValuePicker
  //       value={v}
  //       setValue={sv}
  //       options={[
  //         "Interpolation (Medium).txt",
  //         "No Interpolation.txt",
  //         "Scanlines (Medium).txt",
  //         "Scanlines (Strong).txt",
  //         "Scanlines (Weak).txt",
  //       ]}
  //       formatOption={(option) => option.replace(/\.[^/.]+$/, "")}
  //     />
  //   </FormControl>
  // );
  return <TextOption value={v} setValue={sv} label="Menu font" />;
}

function Logo() {
  const v = useIniSettingsStore((state) => state.logo);
  const sv = useIniSettingsStore((state) => state.setLogo);

  return (
    <BoolOption
      value={v}
      setValue={sv}
      label="Show MiSTer logo on background"
    />
  );
}

export default function OSDMenuSettings() {
  return (
    <>
      <PageHeader title="OSD and Menu" />
      <Stack spacing={3} m={2}>
        <Font />
        <OsdRotate />
        <RbfHideDatecode />
        <BrowseExpand />
        <OSDTimeout />
        <VideoOff />
        <Logo />
        <LogFileEntry />
        <MenuPal />
      </Stack>
      <SaveButton />
    </>
  );
}
