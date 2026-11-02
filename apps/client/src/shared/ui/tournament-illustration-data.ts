import type { AnimationObject } from "lottie-react-native";

export type TournamentIllustrationVariant = "trophy" | "bracket" | "whistle" | "team" | "welcome";
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
  const duration = variant === "welcome" ? 78 : variant === "bracket" ? 60 : 48;
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
  } else if (variant === "whistle") {
    layer(
      "Air · upper",
      [
        path([
          [167, 64],
          [190, 54],
        ]),
        stroke(palette.accent, 3),
        trim(12, 28),
      ],
      {
        opacity: tween([0], [100], 10, 22),
        position: tween([0, 3, 0], [0, 0, 0], 12, 34),
      },
    );
    layer(
      "Air · lower",
      [
        path([
          [173, 88],
          [199, 88],
        ]),
        stroke(palette.accent, 3),
        trim(16, 32),
      ],
      { opacity: tween([0], [100], 14, 26) },
    );
    layer(
      "Whistle",
      [
        path(
          [
            [99, 72],
            [156, 72],
            [156, 94],
            [131, 94],
            [120, 120],
            [91, 120],
            [77, 105],
            [77, 89],
            [99, 72],
          ],
          true,
        ),
        stroke(palette.accent, 4),
        trim(0, 26),
      ],
      {
        anchor: [128, 96],
        position: tween([128, 104, 0], [128, 96, 0], 0, 30),
        scale: tween([88, 88, 100], [100, 100, 100], 0, 30),
      },
    );
    layer("Whistle opening", [ellipse([103, 98], [17, 17]), stroke(palette.ink, 3), trim(10, 30)]);
    layer("Mouthpiece", [
      path([
        [136, 73],
        [136, 92],
      ]),
      stroke(palette.ink, 2),
      trim(12, 32),
    ]);
    layer("Whistle halo", [ellipse([128, 96], [132, 120]), fill(palette.subtle)]);
  } else if (variant === "team") {
    const shirt = (
      name: string,
      x: number,
      y: number,
      scale: number,
      hex: string,
      arriving = false,
    ) => {
      layer(
        name,
        [
          path(
            [
              [-14, -28],
              [-29, -19],
              [-22, -5],
              [-15, -9],
              [-15, 27],
              [15, 27],
              [15, -9],
              [22, -5],
              [29, -19],
              [14, -28],
              [7, -20],
              [-7, -20],
            ],
            true,
          ),
          stroke(hex, 3),
          trim(arriving ? 4 : 0, arriving ? 28 : 18),
        ],
        {
          position: arriving ? tween([x, y - 32, 0], [x, y, 0], 4, 34) : still([x, y, 0]),
          scale: arriving
            ? tween([scale * 0.8, scale * 0.8, 100], [scale, scale, 100], 4, 34)
            : still([scale, scale, 100]),
          opacity: tween([0], [100], arriving ? 4 : 0, arriving ? 16 : 12),
        },
      );
    };
    shirt("New teammate", 128, 98, 110, palette.accent, true);
    layer("Team check", [
      path([
        [119, 103],
        [126, 110],
        [139, 94],
      ]),
      stroke(palette.accent, 3),
      trim(26, 42),
    ]);
    shirt("Left teammate", 68, 105, 74, palette.ink);
    shirt("Right teammate", 188, 105, 74, palette.ink);
    layer("Team ground", [ellipse([128, 147], [158, 6]), fill(palette.subtle)]);
    layer("Team halo", [ellipse([128, 96], [116, 116]), fill(palette.subtle)]);
  } else {
    if (variant === "welcome") {
      layer(
        "Ball becomes forward mark",
        [
          ellipse([0, 0], [14, 14]),
          fill(palette.subtle),
          stroke(palette.accent, 2),
          path(
            [
              [-3, -4],
              [3, -4],
              [5, 1],
              [0, 5],
              [-5, 1],
            ],
            true,
          ),
          stroke(palette.ink, 1),
        ],
        {
          position: {
            a: 1,
            k: [
              { t: 0, s: [37, 49, 0], e: [77, 49, 0], ...ease },
              { t: 14, s: [77, 49, 0], e: [77, 93, 0], ...ease },
              { t: 30, s: [77, 93, 0], e: [128, 93, 0], ...ease },
              { t: 50, s: [128, 93, 0] },
            ],
          },
          scale: tween([100, 100, 100], [0, 0, 100], 50, 58),
        },
      );
      layer(
        "FastTourney forward",
        [
          path(
            [
              [118, 78],
              [129, 88],
              [118, 98],
            ],
            true,
          ),
          path(
            [
              [129, 78],
              [140, 88],
              [129, 98],
            ],
            true,
          ),
          fill(palette.accent),
        ],
        { opacity: tween([0], [100], 52, 66) },
      );
    }
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
