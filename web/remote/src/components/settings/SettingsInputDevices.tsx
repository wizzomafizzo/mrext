import FormHelperText from "@mui/material/FormHelperText";
import Stack from "@mui/material/Stack";
import {
  BoolOption,
  NumberOption,
  PageHeader,
  SaveButton,
  SectionHeader,
  SimpleSelectOption,
  TextOption,
  ToggleableNumberSliderOption,
} from "./SettingsCommon";

import { useIniSettingsStore } from "../../lib/ini";

function BtAutoDisconnect() {
  const v = useIniSettingsStore((state) => state.btAutoDisconnect);
  const sv = useIniSettingsStore((state) => state.setBtAutoDisconnect);

  return (
    <ToggleableNumberSliderOption
      label="Bluetooth auto disconnect"
      value={v}
      setValue={sv}
      min={1}
      max={60}
      defaultValue={"1"}
      step={1}
      suffix="minutes"
      helpText="Automatically disconnects Bluetooth controllers after a period of inactivity. Useful for controllers which do not automatically do this themselves."
    />
  );
}

function BtResetBeforePair() {
  const v = useIniSettingsStore((state) => state.btResetBeforePair);
  const sv = useIniSettingsStore((state) => state.setBtResetBeforePair);

  return (
    <BoolOption
      value={v}
      setValue={sv}
      label="Reset Bluetooth before pairing"
      helpText="May fix issues where Bluetooth device will not pair with dongle."
    />
  );
}

function MouseThrottle() {
  const v = useIniSettingsStore((state) => state.mouseThrottle);
  const sv = useIniSettingsStore((state) => state.setMouseThrottle);

  return (
    <ToggleableNumberSliderOption
      label="Mouse speed throttling"
      value={v}
      setValue={sv}
      min={1}
      max={100}
      defaultValue={"10"}
      step={1}
      helpText="Divides the mouse speed by given number. Useful for very sensitive mice."
    />
  );
}

function ResetCombo() {
  const v = useIniSettingsStore((state) => state.resetCombo);
  const sv = useIniSettingsStore((state) => state.setResetCombo);

  return (
    <SimpleSelectOption
      value={v}
      setValue={sv}
      options={[
        "LCtrl+LAlt+RAlt (Keyrah: LCtrl+LGui+Ralt)",
        "LCtrl+LGUI+RGUI",
        "LCtrl+LAlt+Del",
        "LCtrl+LAlt+RAlt (Keyrah: LCtrl+LAlt+Ralt)",
      ]}
      label="USER button keyboard combination"
    />
  );
}

function PlayerControllers() {
  const v1 = useIniSettingsStore((state) => state.player1Controller);
  const sv1 = useIniSettingsStore((state) => state.setPlayer1Controller);
  const v2 = useIniSettingsStore((state) => state.player2Controller);
  const sv2 = useIniSettingsStore((state) => state.setPlayer2Controller);
  const v3 = useIniSettingsStore((state) => state.player3Controller);
  const sv3 = useIniSettingsStore((state) => state.setPlayer3Controller);
  const v4 = useIniSettingsStore((state) => state.player4Controller);
  const sv4 = useIniSettingsStore((state) => state.setPlayer4Controller);
  const v5 = useIniSettingsStore((state) => state.player5Controller);
  const sv5 = useIniSettingsStore((state) => state.setPlayer5Controller);
  const v6 = useIniSettingsStore((state) => state.player6Controller);
  const sv6 = useIniSettingsStore((state) => state.setPlayer6Controller);

  return (
    <>
      <TextOption value={v1} setValue={sv1} label="Player 1 controller" />
      <TextOption value={v2} setValue={sv2} label="Player 2 controller" />
      <TextOption value={v3} setValue={sv3} label="Player 3 controller" />
      <TextOption value={v4} setValue={sv4} label="Player 4 controller" />
      <TextOption value={v5} setValue={sv5} label="Player 5 controller" />
      <TextOption value={v6} setValue={sv6} label="Player 6 controller" />
      <FormHelperText>Assign a specific USB device to a player.</FormHelperText>
    </>
  );
}

function SniperMode() {
  const v = useIniSettingsStore((state) => state.sniperMode);
  const sv = useIniSettingsStore((state) => state.setSniperMode);

  return (
    <SimpleSelectOption
      label="Sniper mode speeds"
      value={v}
      setValue={sv}
      options={[
        "Slow movement in sniper mode",
        "Slow movement outside sniper mode",
      ]}
    />
  );
}

function SpinnerThrottle() {
  const v = useIniSettingsStore((state) => state.spinnerThrottle);
  const sv = useIniSettingsStore((state) => state.setSpinnerThrottle);

  return (
    <NumberOption
      value={v}
      setValue={sv}
      label="Spinner speed throttling"
      helpText="Base value of 100 gives 1 spinner step per tick. Higher values will slow down spinner speed, lower values will speed it up, and negative values will invert the spinner direction."
      min={-10000}
      max={10000}
      width={120}
      defaultValue={"100"}
      disabledValue={""}
    />
  );
}

function SpinnerAxis() {
  const v = useIniSettingsStore((state) => state.spinnerAxis);
  const sv = useIniSettingsStore((state) => state.setSpinnerAxis);

  return (
    <SimpleSelectOption
      label="Spinner axis"
      value={v}
      setValue={sv}
      options={["X axis", "Y axis", "Wheel"]}
    />
  );
}

function GamepadDefaults() {
  const v = useIniSettingsStore((state) => state.gamepadDefaults);
  const sv = useIniSettingsStore((state) => state.setGamepadDefaults);

  return (
    <SimpleSelectOption
      label="Default internal gamepad mapping"
      value={v}
      setValue={sv}
      options={["Name mapping", "Positional mapping"]}
      helpText={
        v == "0"
          ? "e.g. A button in SNES core = A button on controller regardless of position on pad."
          : "e.g. A button in SNES core = East button on controller regardless of button name."
      }
    />
  );
}

function WheelForce() {
  const v = useIniSettingsStore((state) => state.wheelForce);
  const sv = useIniSettingsStore((state) => state.setWheelForce);

  return (
    <NumberOption
      value={v}
      setValue={sv}
      label="Wheel centering force"
      min={0}
      max={100}
      defaultValue={"50"}
      disabledValue={""}
    />
  );
}

function WheelRange() {
  const v = useIniSettingsStore((state) => state.wheelRange);
  const sv = useIniSettingsStore((state) => state.setWheelRange);

  // TODO: no idea what the ranges should be for this
  return (
    <NumberOption
      value={v}
      setValue={sv}
      label="Wheel steering angle range"
      min={-10000}
      max={10000}
      defaultValue={"200"}
      disabledValue={""}
      helpText="Supported ranges depends on specific wheel model. If not set, then default range (depending on driver) is used."
    />
  );
}

function SpinnerVid() {
  const v = useIniSettingsStore((state) => state.spinnerVid);
  const sv = useIniSettingsStore((state) => state.setSpinnerVid);

  return <TextOption value={v} setValue={sv} label="Spinner VID" />;
}

function SpinnerPid() {
  const v = useIniSettingsStore((state) => state.spinnerPid);
  const sv = useIniSettingsStore((state) => state.setSpinnerPid);

  return <TextOption value={v} setValue={sv} label="Spinner PID" />;
}

function KeyrahMode() {
  const v = useIniSettingsStore((state) => state.keyrahMode);
  const sv = useIniSettingsStore((state) => state.setKeyrahMode);

  return <TextOption value={v} setValue={sv} label="Keyrah mode" />;
}

function JammaVid() {
  const v = useIniSettingsStore((state) => state.jammaVid);
  const sv = useIniSettingsStore((state) => state.setJammaVid);

  return <TextOption value={v} setValue={sv} label="Jamma VID" />;
}

function JammaPid() {
  const v = useIniSettingsStore((state) => state.jammaPid);
  const sv = useIniSettingsStore((state) => state.setJammaPid);

  return <TextOption value={v} setValue={sv} label="Jamma PID" />;
}

function NoMergeVid() {
  const v = useIniSettingsStore((state) => state.noMergeVid);
  const sv = useIniSettingsStore((state) => state.setNoMergeVid);

  return <TextOption value={v} setValue={sv} label="No merge VID" />;
}

function NoMergePid() {
  const v = useIniSettingsStore((state) => state.noMergePid);
  const sv = useIniSettingsStore((state) => state.setNoMergePid);

  return <TextOption value={v} setValue={sv} label="No merge PID" />;
}

function NoMergeVidPid() {
  const v = useIniSettingsStore((state) => state.noMergeVidPid);
  const sv = useIniSettingsStore((state) => state.setNoMergeVidPid);

  // TODO: this option supports shadowed entries in the ini

  // return <TextOption value={v} setValue={sv} label="No merge VID:PID" />;
  return <></>;
}

function AutoFire() {
  const v = useIniSettingsStore((state) => state.disableAutoFire);
  const sv = useIniSettingsStore((state) => state.setDisableAutoFire);

  return <BoolOption value={v} setValue={sv} label="Enable autofire" invert />;
}

export default function InputDevices() {
  return (
    <>
      <PageHeader title="Input Devices" />
      <Stack spacing={3} m={2}>
        <BtAutoDisconnect />
        <BtResetBeforePair />
        <AutoFire />
        <SniperMode />

        <SectionHeader text="Keyboard and mouse" />
        <MouseThrottle />
        <ResetCombo />

        <SectionHeader text="Wheels and spinners" />
        <WheelForce />
        <WheelRange />
        <SpinnerThrottle />
        <SpinnerAxis />

        <SectionHeader text="Advanced" />
        <PlayerControllers />
        <GamepadDefaults />
        <SpinnerVid />
        <SpinnerPid />
        <KeyrahMode />
        <JammaVid />
        <JammaPid />
        <NoMergeVid />
        <NoMergePid />
        <NoMergeVidPid />
      </Stack>
      <SaveButton />
    </>
  );
}
