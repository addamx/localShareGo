import { defineComponent, h, mergeProps, type VNode } from "vue";

function createIcon(name: string, renderChildren: () => VNode[]) {
  return defineComponent({
    name,
    setup(_, { attrs }) {
      return () =>
        h(
          "svg",
          mergeProps(
            {
              viewBox: "0 0 24 24",
              fill: "none",
              stroke: "currentColor",
              "stroke-width": "1.7",
              "stroke-linecap": "round",
              "stroke-linejoin": "round",
            },
            attrs,
          ),
          renderChildren(),
        );
    },
  });
}

export const ShareLogoIcon = createIcon("ShareLogoIcon", () => [
  h("circle", { cx: "6.5", cy: "12", r: "2.2" }),
  h("circle", { cx: "17.5", cy: "6.5", r: "2.2" }),
  h("circle", { cx: "17.5", cy: "17.5", r: "2.2" }),
  h("path", { d: "M8.6 10.9 15.2 7.6" }),
  h("path", { d: "M8.6 13.1 15.2 16.4" }),
]);

export const GlobeIcon = createIcon("GlobeIcon", () => [
  h("circle", { cx: "12", cy: "12", r: "8" }),
  h("path", { d: "M4 12h16" }),
  h("path", { d: "M12 4c2.5 2.2 4 5 4 8s-1.5 5.8-4 8" }),
  h("path", { d: "M12 4c-2.5 2.2-4 5-4 8s1.5 5.8 4 8" }),
]);

export const HelpIcon = createIcon("HelpIcon", () => [
  h("circle", { cx: "12", cy: "12", r: "8" }),
  h("path", { d: "M9.4 9.2a2.7 2.7 0 1 1 5.2.9c-.6 1.4-2.3 1.8-2.6 3.5" }),
  h("path", { d: "M12 17h.01" }),
]);

export const RotateIcon = createIcon("RotateIcon", () => [
  h("path", { d: "M18 4v4h-4" }),
  h("path", { d: "M6 20v-4h4" }),
  h("path", { d: "M20 12a8 8 0 0 0-13.6-5.8L4 8" }),
  h("path", { d: "M4 12a8 8 0 0 0 13.6 5.8L20 16" }),
]);

export const CopyIcon = createIcon("CopyIcon", () => [
  h("rect", { x: "9", y: "9", width: "10", height: "11", rx: "1.6" }),
  h("path", { d: "M6 15H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h8a2 2 0 0 1 2 2v1" }),
]);

export const OpenIcon = createIcon("OpenIcon", () => [
  h("path", { d: "M14 5h5v5" }),
  h("path", { d: "M10 14 19 5" }),
  h("path", { d: "M19 13v5a1 1 0 0 1-1 1H6a1 1 0 0 1-1-1V6a1 1 0 0 1 1-1h5" }),
]);

export const RouteIcon = createIcon("RouteIcon", () => [
  h("circle", { cx: "6", cy: "6", r: "2" }),
  h("circle", { cx: "18", cy: "10", r: "2" }),
  h("circle", { cx: "9", cy: "18", r: "2" }),
  h("path", { d: "M7.8 7.1 16.2 8.9" }),
  h("path", { d: "M7.4 7.8 8.4 16.1" }),
]);

export const DiagnosticsIcon = createIcon("DiagnosticsIcon", () => [
  h("path", { d: "M3 12h4l2.2-4.2 4 8 2.1-4H21" }),
]);

export const MoreIcon = createIcon("MoreIcon", () => [
  h("circle", { cx: "6", cy: "12", r: "1.4", fill: "currentColor", stroke: "none" }),
  h("circle", { cx: "12", cy: "12", r: "1.4", fill: "currentColor", stroke: "none" }),
  h("circle", { cx: "18", cy: "12", r: "1.4", fill: "currentColor", stroke: "none" }),
]);

export const SearchIcon = createIcon("SearchIcon", () => [
  h("circle", { cx: "11", cy: "11", r: "6" }),
  h("path", { d: "m16 16 4 4" }),
]);

export const EyeIcon = createIcon("EyeIcon", () => [
  h("path", { d: "M2.8 12s3.2-5.5 9.2-5.5 9.2 5.5 9.2 5.5-3.2 5.5-9.2 5.5S2.8 12 2.8 12Z" }),
  h("circle", { cx: "12", cy: "12", r: "2.5" }),
]);

export const DeleteIcon = createIcon("DeleteIcon", () => [
  h("path", { d: "M4 7h16" }),
  h("path", { d: "M9 7V4.8A1.8 1.8 0 0 1 10.8 3h2.4A1.8 1.8 0 0 1 15 4.8V7" }),
  h("path", { d: "M8 7v12" }),
  h("path", { d: "M12 7v12" }),
  h("path", { d: "M16 7v12" }),
]);

export const CloseIcon = createIcon("CloseIcon", () => [
  h("path", { d: "M6 6 18 18" }),
  h("path", { d: "M18 6 6 18" }),
]);
