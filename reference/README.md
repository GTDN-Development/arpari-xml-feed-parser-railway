# Reference data

This directory stores normalized external inputs used for feed mapping decisions.

- `shoptet-categories.csv` is a compact Shoptet category export derived from
  `/Users/fanda/Downloads/categories (1).csv`.
- The normalized file intentionally keeps only stable category fields:
  `id`, `parent_id`, `title`, `url`, `visible`, and full category `path`.
- Rich HTML descriptions from the Shoptet export are not stored here because
  feed transforms only need category identity and hierarchy.
