import { Children, isValidElement } from "react";

import { space, webLayout } from "@tournaments-manager/design-tokens";

import type { ResponsiveColumnsProps } from "./responsive-columns";

// Each Card keeps its 20 px margins. Half-margin compensation shares the gutter
// between neighboring cells, preserving 20 px both outside and between cards.
const css = `
  .tm-columns { display: grid; grid-template-columns: minmax(0, 1fr); }
  .tm-column { min-width: 0; }
  @media (min-width: ${webLayout.columnsMinWidth}px) {
    .tm-columns { grid-template-columns: repeat(2, minmax(0, 1fr)); margin-inline: ${space[5] / 2}px; }
    .tm-column { margin-inline: -${space[5] / 2}px; }
    .tm-columns-fill-last > .tm-column:last-child:nth-child(odd) { grid-column: 1 / -1; }
  }
`;

export function ResponsiveColumns({
  children,
  gap = space[5],
  fullWidthLast = false,
}: ResponsiveColumnsProps) {
  return (
    <>
      <style>{css}</style>
      <div
        className={`tm-columns${fullWidthLast ? " tm-columns-fill-last" : ""}`}
        style={{ rowGap: gap }}
      >
        {Children.toArray(children).map((child, index) => (
          <div className="tm-column" key={isValidElement(child) ? child.key : index}>
            {child}
          </div>
        ))}
      </div>
    </>
  );
}
