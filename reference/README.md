# Reference data

This directory stores normalized external inputs used for feed mapping decisions.

- `shoptet-categories.csv` is a compact Shoptet category export derived from
  `/Users/fanda/Downloads/categories.csv`.
- The normalized file intentionally keeps only stable category fields:
  `id`, `parent_id`, `title`, `url`, `visible`, and full category `path`.
- Rich HTML descriptions from the Shoptet export are not stored here because
  feed transforms only need category identity and hierarchy.
- `drevocal/drevocal-b2b-feed-dokumentace-2026-05.pdf` is the supplier-provided
  Dřevočal B2B XML feed documentation. It defines the `feed-b2b.xml` source,
  `ITEMGROUP_ID` product grouping, `ITEM_ID` variant SKU, and variant
  parameters `Rozměr`, `Výška`, and `Potah`.
