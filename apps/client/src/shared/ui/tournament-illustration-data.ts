import type { AnimationObject } from "lottie-react-native";

export type TournamentIllustrationVariant = "trophy" | "bracket";
type Palette = { accent: string; ink: string; subtle: string };
type Point = [number, number];

const still = (k: unknown) => ({ a: 0, k });
const ease = { i: { x: [0.65], y: [1] }, o: { x: [0.25], y: [0] } };
const tween = (from: number[], to: number[], start: number, end: number) => ({
  a: 1,
  k: [
    { t: start, s: from, e: to, ...ease },
    { t: end, s: to },
  ],
});
const rgba = (hex: string) =>
  [1, 3, 5].map((offset) => parseInt(hex.slice(offset, offset + 2), 16) / 255).concat(1);
const stroke = (hex: string, width = 3) => ({
  ty: "st",
  c: still(rgba(hex)),
  o: still(100),
  w: still(width),
  lc: 2,
  lj: 2,
});
const fill = (hex: string) => ({ ty: "fl", c: still(rgba(hex)), o: still(100), r: 1 });
const trim = (start: number, end: number) => ({
  ty: "tm",
  s: still(0),
  e: tween([0], [100], start, end),
  o: still(0),
  m: 1,
});
const path = (vertices: Point[], closed = false, incoming?: Point[], outgoing?: Point[]) => ({
  ty: "sh",
  ks: still({
    v: vertices,
    i: incoming ?? vertices.map(() => [0, 0]),
    o: outgoing ?? vertices.map(() => [0, 0]),
    c: closed,
  }),
});
const rect = (position: Point, size: Point, radius = 6) => ({
  ty: "rc",
  p: still(position),
  s: still(size),
  r: still(radius),
  d: 1,
});
const ellipse = (position: Point, size: Point) => ({
  ty: "el",
  p: still(position),
  s: still(size),
  d: 1,
});

/** Original Grada artwork: only paths, fills and transforms; no external assets or expressions. */
export function createTournamentIllustration(
  variant: TournamentIllustrationVariant,
  palette: Palette,
): AnimationObject {
  const layers: AnimationObject["layers"] = [];
  const duration = variant === "trophy" ? 48 : 60;
  const layer = (
    name: string,
    shapes: unknown[],
    options: { scale?: unknown; position?: unknown; opacity?: unknown; anchor?: Point } = {},
  ) => {
    layers.push({
      ddd: 0,
      ind: layers.length + 1,
      ty: 4,
      nm: name,
      sr: 1,
      ks: {
        o: options.opacity ?? still(100),
        r: still(0),
        p: options.position ?? still([0, 0, 0]),
        a: still([...(options.anchor ?? [0, 0]), 0]),
        s: options.scale ?? still([100, 100, 100]),
      },
      ao: 0,
      shapes,
      ip: 0,
      op: duration,
      st: 0,
      bm: 0,
    });
  };

  const cup = (small: boolean) => {
    // The bowl uses cubic tangents, shared by the celebration and bracket.
    const bowl = path(
      [
        [101, 56],
        [155, 56],
        [128, 111],
      ],
      true,
      [
        [0, 25],
        [0, 0],
        [21, 0],
      ],
      [
        [0, 0],
        [0, 25],
        [-21, 0],
      ],
    );
    const handles = [
      path(
        [
          [101, 64],
          [83, 64],
          [105, 94],
        ],
        false,
        [
          [0, 0],
          [0, 0],
          [-21, 0],
        ],
        [
          [0, 0],
          [0, 24],
          [0, 0],
        ],
      ),
      path(
        [
          [155, 64],
          [173, 64],
          [151, 94],
        ],
        false,
        [
          [0, 0],
          [0, 0],
          [21, 0],
        ],
        [
          [0, 0],
          [0, 24],
          [0, 0],
        ],
      ),
    ];
    const start = small ? 28 : 0;
    layer(
      "Cup · Grada",
      [
        bowl,
        ...handles,
        path([
          [128, 112],
          [128, 132],
        ]),
        path([
          [109, 135],
          [147, 135],
        ]),
        stroke(palette.accent, 4),
        trim(start, start + 24),
      ],
      {
        anchor: [128, 96],
        position: still([128, 96, 0]),
        scale: small
          ? tween([0, 0, 100], [68, 68, 100], 26, 44)
          : tween([86, 86, 100], [100, 100, 100], 0, 30),
      },
    );
    layer(
      "Cup rim",
      [
        path([
          [108, 63],
          [148, 63],
        ]),
        stroke(palette.accent, 2),
        trim(start + 8, start + 26),
      ],
      {
        anchor: [128, 96],
        position: still([128, 96, 0]),
        scale: small
          ? tween([0, 0, 100], [68, 68, 100], 26, 44)
          : tween([86, 86, 100], [100, 100, 100], 0, 30),
      },
    );
  };

  if (variant === "trophy") {
    for (const [index, [x, y]] of (
      [
        [65, 53],
        [187, 43],
        [192, 117],
        [62, 127],
      ] as Point[]
    ).entries()) {
      layer(
        `Spark ${index + 1}`,
        [
          path(
            [
              [0, -5],
              [2, -2],
              [5, 0],
              [2, 2],
              [0, 5],
              [-2, 2],
              [-5, 0],
              [-2, -2],
            ],
            true,
          ),
          fill(index % 2 ? palette.ink : palette.accent),
        ],
        {
          position: tween([128 + (x - 128) * 0.7, 96 + (y - 96) * 0.7, 0], [x, y, 0], 12, 42),
          scale: {
            a: 1,
            k: [
              { t: 9 + index * 2, s: [0, 0, 100], e: [100, 100, 100], ...ease },
              { t: 22 + index * 2, s: [100, 100, 100], e: [0, 0, 100], ...ease },
              { t: 46, s: [0, 0, 100] },
            ],
          },
        },
      );
    }
    cup(false);
    layer("Ground", [ellipse([128, 154], [64, 5]), fill(palette.subtle)]);
    layer("Halo", [ellipse([128, 92], [128, 128]), fill(palette.subtle)], {
      scale: tween([88, 88, 100], [100, 100, 100], 0, 34),
      anchor: [128, 92],
      position: still([128, 92, 0]),
    });
  } else {
    cup(true);
    const positions: Point[] = [
      [37, 49],
      [37, 137],
      [219, 49],
      [219, 137],
    ];
    positions.forEach(([x, y], index) => {
      layer(`Team ${index + 1}`, [rect([x, y], [44, 24]), stroke(palette.ink, 2)], {
        opacity: tween([0], [100], index * 4, index * 4 + 12),
      });
      layer(
        `Team mark ${index + 1}`,
        [
          ellipse([x - 11, y], [5, 5]),
          fill(palette.accent),
          path([
            [x - 3, y],
            [x + 12, y],
          ]),
          stroke(palette.accent, 2),
        ],
        { opacity: tween([0], [100], index * 4 + 4, index * 4 + 16) },
      );
    });
    layer("Left bracket", [
      path([
        [60, 49],
        [77, 49],
        [77, 137],
        [60, 137],
      ]),
      path([
        [77, 93],
        [95, 93],
      ]),
      stroke(palette.accent, 2),
      trim(8, 32),
    ]);
    layer("Right bracket", [
      path([
        [196, 49],
        [179, 49],
        [179, 137],
        [196, 137],
      ]),
      path([
        [179, 93],
        [161, 93],
      ]),
      stroke(palette.accent, 2),
      trim(14, 38),
    ]);
    layer("Center stage", [ellipse([128, 96], [74, 74]), fill(palette.subtle)]);
  }
  return {
    v: "5.7.4",
    fr: 30,
    ip: 0,
    op: duration,
    w: 256,
    h: 192,
    nm: `FastTourney · ${variant}`,
    ddd: 0,
    assets: [],
    layers,
    markers: [],
  };
}
