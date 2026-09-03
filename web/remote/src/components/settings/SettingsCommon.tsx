import { useEffect, useState } from "react";

import Button from "@mui/material/Button";
import Stack from "@mui/material/Stack";
import Typography from "@mui/material/Typography";
import ArrowBackIcon from "@mui/icons-material/ArrowBack";
import Dialog from "@mui/material/Dialog";
import List from "@mui/material/List";
import Box from "@mui/material/Box";
import ListItem from "@mui/material/ListItem";
import ListItemText from "@mui/material/ListItemText";
import FormControl from "@mui/material/FormControl";
import FormControlLabel from "@mui/material/FormControlLabel";
import Checkbox from "@mui/material/Checkbox";
import FormHelperText from "@mui/material/FormHelperText";
import InputLabel from "@mui/material/InputLabel";
import Select from "@mui/material/Select";
import MenuItem from "@mui/material/MenuItem";
import TextField from "@mui/material/TextField";
import FormLabel from "@mui/material/FormLabel";
import Slider from "@mui/material/Slider";
import Grid from "@mui/material/Grid";
import Paper from "@mui/material/Paper";

import { useUIStateStore, SettingsPageId } from "../../lib/store";
import { saveMisterIni, useIniSettingsStore } from "../../lib/ini";
import { ControlApi } from "../../lib/api";
import { ListInisPayload } from "../../lib/models";

export function activeIniId(inis: ListInisPayload): number {
  if (inis.active === 0) {
    return 1;
  } else {
    return inis.active;
  }
}

export function PageHeader(props: { title: string; noRevert?: boolean }) {
  const setActiveSettingsPage = useUIStateStore(
    (state) => state.setActiveSettingsPage
  );
  const modified = useIniSettingsStore((state) => state.modified);
  const revertChanges = useIniSettingsStore((state) => state.revertChanges);
  const resetModified = useIniSettingsStore((state) => state.resetModified);

  const handleRevert = () => {
    resetModified();
    revertChanges();
  };

  return (
    <>
      <Paper
        sx={{
          boxShadow: 2,
          position: "fixed",
          width: { xs: 1, sm: "calc(100% - 240px)" },
          height: "45px",
          zIndex: 2,
          borderRadius: 0,
        }}
      >
        <Grid container sx={{ alignItems: "center", height: "45px" }}>
          <Grid item xs={3} sx={{ pl: 1 }}>
            <Button
              variant="text"
              onClick={() => setActiveSettingsPage(SettingsPageId.Main)}
              startIcon={<ArrowBackIcon />}
            >
              Back
            </Button>
          </Grid>
          <Grid item xs={6} sx={{ textAlign: "center" }}>
            <Typography variant="h6">{props.title}</Typography>
          </Grid>
          <Grid item xs={3} sx={{ textAlign: "right" }}>
            {modified.length > 0 && !props.noRevert && (
              <Button
                variant="text"
                color="error"
                onClick={() => handleRevert()}
              >
                Revert
              </Button>
            )}
          </Grid>
        </Grid>
      </Paper>
      <Box sx={{ height: "45px" }}></Box>
    </>
  );
}

export function SaveButton() {
  const store = useIniSettingsStore();
  const modified = useIniSettingsStore((state) => state.modified);

  return (
    <>
      <Paper
        sx={{
          boxShadow: 2,
          position: "fixed",
          bottom: 0,
          width: { xs: 1, sm: "calc(100% - 240px)" },
          height: "50px",
          zIndex: 2,
          borderRadius: 0,
        }}
      >
        <Grid container sx={{ alignItems: "center", height: "50px" }}>
          <Grid item xs={12} sx={{ p: 1, pt: 0.5, pb: 0.5 }}>
            <Button
              variant="contained"
              onClick={() => {
                const api = new ControlApi();
                api.listMisterInis().then((inis) => {
                  saveMisterIni(activeIniId(inis), store).catch((err) =>
                    console.error(err)
                  );
                });
              }}
              disabled={modified.length === 0}
              color="success"
              fullWidth
            >
              Save
            </Button>
          </Grid>
        </Grid>
      </Paper>
      <Box sx={{ height: "50px" }}></Box>
    </>
  );
}

export function ValuePicker(props: {
  value: string;
  setValue: (value: string) => void;
  options: string[];
  formatOption: (option: string) => string;
}) {
  const unsetValue = "<unset>";
  const [open, setOpen] = useState(false);

  return (
    <FormControl>
      <Stack
        spacing={2}
        direction="row"
        justifyContent="space-between"
        alignItems="center"
      >
        <Typography>
          {props.value !== "" ? props.formatOption(props.value) : unsetValue}
        </Typography>
        <Button size="small" onClick={() => setOpen(true)}>
          Browse...
        </Button>
      </Stack>
      <Dialog onClose={() => setOpen(false)} open={open}>
        <List>
          {props.options.map((option) => (
            <ListItem
              button
              onClick={() => {
                props.setValue(option);
                setOpen(false);
              }}
              key={option}
            >
              <ListItemText primary={props.formatOption(option)} />
            </ListItem>
          ))}
        </List>
      </Dialog>
    </FormControl>
  );
}

export function BoolOption(props: {
  value: string;
  setValue: (value: string) => void;
  label: string;
  helpText?: string;
  invert?: boolean;
}) {
  return (
    <FormControl>
      <FormControlLabel
        control={
          <Checkbox
            checked={props.invert ? props.value == "0" : props.value == "1"}
            onChange={(e) => {
              props.setValue(
                props.invert
                  ? e.target.checked
                    ? "0"
                    : "1"
                  : e.target.checked
                  ? "1"
                  : "0"
              );
            }}
          />
        }
        label={props.label}
      />
      {props.helpText && <FormHelperText>{props.helpText}</FormHelperText>}
    </FormControl>
  );
}

export function TextOption(props: {
  value: string;
  setValue: (value: string) => void;
  label: string;
  helpText?: string;
}) {
  return (
    <FormControl>
      {/*<InputLabel>{props.label}</InputLabel>*/}
      <TextField
        value={props.value}
        onChange={(e) => props.setValue(e.target.value)}
        label={props.label}
        inputProps={{ maxLength: 1023 }}
      />
      {props.helpText && <FormHelperText>{props.helpText}</FormHelperText>}
    </FormControl>
  );
}

export function SelectOption(props: {
  value: string;
  setValue: (value: string) => void;
  optionLabels: string[];
  optionValues: string[];
  helpText?: string[] | string;
  label: string;
}) {
  const value = props.value === "" ? "__none__" : props.value;

  return (
    <FormControl>
      <InputLabel>{props.label}</InputLabel>
      <Select
        value={value}
        onChange={(e) => {
          if (e.target.value === "__none__") {
            props.setValue("");
          } else {
            props.setValue(e.target.value);
          }
        }}
        label={props.label}
        defaultValue={props.optionValues[0]}
      >
        {props.optionValues.map((option, i) => (
          <MenuItem key={option} value={option === "" ? "__none__" : option}>
            {props.optionLabels[i]}
          </MenuItem>
        ))}
      </Select>
      {Array.isArray(props.helpText) &&
      props.helpText.length === props.optionValues.length ? (
        <FormHelperText>
          {
            props.helpText[
              props.optionValues.indexOf(value === "__none__" ? "" : value)
            ]
          }
        </FormHelperText>
      ) : props.helpText !== undefined ? (
        <FormHelperText>{props.helpText}</FormHelperText>
      ) : null}
    </FormControl>
  );
}

export function SimpleSelectOption(props: {
  value: string;
  setValue: (value: string) => void;
  options: string[];
  helpText?: string[] | string;
  label: string;
}) {
  const vals = props.options.map((option, i) => i.toString());
  return (
    <SelectOption
      value={props.value}
      setValue={props.setValue}
      optionLabels={props.options}
      optionValues={vals}
      helpText={props.helpText}
      label={props.label}
    />
  );
}

export function NumberOption(props: {
  value: string;
  setValue: (value: string) => void;
  label: string;
  helpText?: string;
  min: number;
  max: number;
  defaultValue: string;
  disabledValue: string;
  width?: number;
  suffix?: string;
}) {
  const [enabled, setEnabled] = useState(props.value !== props.disabledValue);

  useEffect(() => {
    setEnabled(props.value !== props.disabledValue);
  }, [props.value]);

  return (
    <FormControl>
      <Stack
        spacing={2}
        direction="row"
        alignItems="center"
        justifyContent="space-between"
      >
        <FormControlLabel
          control={
            <Checkbox
              checked={enabled}
              onChange={(e) => {
                if (e.target.checked) {
                  props.setValue(props.defaultValue);
                  setEnabled(true);
                } else {
                  props.setValue(props.disabledValue);
                  setEnabled(false);
                }
              }}
            />
          }
          label={props.label}
        />
        {enabled ? (
          <Stack
            direction="row"
            spacing={1}
            sx={{
              alignItems: "center",
            }}
          >
            <TextField
              type="number"
              inputProps={{
                inputMode: "numeric",
                pattern: "[0-9]*",
                style: { textAlign: "left" },
              }}
              size="small"
              sx={
                props.width !== undefined
                  ? { width: props.width + "px" }
                  : { width: "100px" }
              }
              value={props.value}
              onChange={(e) => {
                const value = Number(e.target.value);
                if (value < props.min) {
                  props.setValue(props.min.toString());
                } else if (value > props.max) {
                  props.setValue(props.max.toString());
                } else {
                  props.setValue(value.toString());
                }
              }}
            />
            {props.suffix ? <div>{props.suffix}</div> : null}
          </Stack>
        ) : null}
      </Stack>
      <FormHelperText>{props.helpText}</FormHelperText>
    </FormControl>
  );
}

export function NumberSliderOption(props: {
  value: string;
  setValue: (value: string) => void;
  label: string;
  helpText?: string;
  min: number;
  max: number;
}) {
  const [internalValue, setInternalValue] = useState(props.value);

  useEffect(() => {
    setInternalValue(props.value);
  }, [props.value]);

  return (
    <FormControl>
      <FormLabel>{props.label}</FormLabel>
      <Stack spacing={1} direction="row" alignItems="center">
        <Slider
          sx={{ ml: 1.5, mr: 1.5 }}
          value={Number(internalValue)}
          min={props.min}
          max={props.max}
          onChange={(e, v) => setInternalValue(v.toString())}
          onChangeCommitted={(e, v) => props.setValue(v.toString())}
        />
        <TextField
          type="number"
          inputProps={{
            inputMode: "numeric",
            pattern: "[0-9]*",
            style: { textAlign: "left" },
          }}
          size="small"
          sx={{ width: 120 }}
          value={internalValue}
          onChange={(e) => {
            const v = Number(e.target.value);
            if (v < props.min) {
              props.setValue(props.min.toString());
            } else if (v > props.max) {
              props.setValue(props.max.toString());
            } else {
              props.setValue(v.toString());
            }
          }}
        />
      </Stack>
      <FormHelperText>{props.helpText}</FormHelperText>
    </FormControl>
  );
}

export function VerticalNumberSliderOption(props: {
  value: string;
  setValue: (value: string) => void;
  label: string;
  min: number;
  max: number;
  step: number;
}) {
  const [internalValue, setInternalValue] = useState(props.value);

  useEffect(() => {
    setInternalValue(props.value);
  }, [props.value]);

  return (
    <FormControl>
      <Stack spacing={2} alignItems="center">
        <FormLabel>{props.label}</FormLabel>
        <Slider
          step={props.step}
          value={Number(internalValue)}
          min={props.min}
          max={props.max}
          sx={{ height: 100 }}
          onChange={(e, v) => setInternalValue(v.toString())}
          onChangeCommitted={(e, v) => props.setValue(v.toString())}
          orientation="vertical"
        />
        <TextField
          type="number"
          inputProps={{
            inputMode: "numeric",
            pattern: "[0-9]*",
            style: { textAlign: "left" },
            step: props.step,
          }}
          size="small"
          sx={{ width: 90 }}
          value={Number(internalValue).toFixed(2)}
          onChange={(e) => {
            const v = Number(e.target.value);
            if (v < props.min) {
              props.setValue(props.min.toString());
            } else if (v > props.max) {
              props.setValue(props.max.toString());
            } else {
              props.setValue(v.toString());
            }
          }}
        />
      </Stack>
    </FormControl>
  );
}

export function ToggleableNumberSliderOption(props: {
  label: string;
  value: string;
  setValue: (value: string) => void;
  helpText?: string;
  min: number;
  max: number;
  step: number;
  defaultValue: string;
  suffix?: string;
}) {
  const [enabled, setEnabled] = useState(props.value !== "0");
  const [internalValue, setInternalValue] = useState(props.value);

  useEffect(() => {
    setEnabled(props.value !== "0");
    setInternalValue(props.value);
  }, [props.value]);

  return (
    <FormControl>
      <FormControlLabel
        control={
          <Checkbox
            checked={enabled}
            onChange={(e) => {
              if (e.target.checked) {
                props.setValue(props.defaultValue);
                setEnabled(true);
              } else {
                props.setValue("0");
                setEnabled(false);
              }
            }}
          />
        }
        label={props.label}
      />
      {enabled ? (
        <Stack spacing={2} direction="row" alignItems="center">
          <Slider
            sx={{ ml: 1.5, mr: 1.5 }}
            value={Number(internalValue)}
            onChangeCommitted={(e, v) => props.setValue(v.toString())}
            onChange={(e, v) => setInternalValue(v.toString())}
            step={props.step}
            min={props.min}
            max={props.max}
          />
          <TextField
            type="number"
            inputProps={{
              inputMode: "numeric",
              pattern: "[0-9]*",
              style: { textAlign: "left" },
            }}
            size="small"
            sx={{ width: "110px" }}
            value={internalValue}
            onChange={(e) => {
              const value = Number(e.target.value);
              if (value < props.min) {
                props.setValue(props.min.toString());
              } else if (value > props.max) {
                props.setValue(props.max.toString());
              } else {
                props.setValue(value.toString());
              }
            }}
          />
          <div>{props.suffix}</div>
        </Stack>
      ) : null}
      <FormHelperText>{props.helpText}</FormHelperText>
    </FormControl>
  );
}

export function SectionHeader(props: { text: string }) {
  return (
    <Typography variant="h6" sx={{ pt: 2 }}>
      {props.text}
    </Typography>
  );
}
