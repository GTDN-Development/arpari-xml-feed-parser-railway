# Feed Task: Sakypaky katalog

## Metadata

- Supplier: Sakypaky
- Generator name: `sakypaky`
- Test generator name: `sakypaky-test`
- Output endpoint: `/feeds/sakypaky.xml`
- Test output endpoint: `/feeds/sakypaky-test.xml`
- Source URL: `https://www.sakypaky.cz/export/b2b_partners_cs.xml`
- Priority: druhá fáze
- Status: implemented, čeká na Shoptet test import
- Last updated: 2026-06-12

## Cíl

Generovat Shoptet XML feed ze Sakypaky katalogového feedu pro nábytkový sortiment a relevantní doplňky.

## Aktuální pravidla

- Zdrojový feed je `https://www.sakypaky.cz/export/b2b_partners_cs.xml`.
- Produkční výstup je `/feeds/sakypaky.xml`.
- Testovací výstup je `/feeds/sakypaky-test.xml`, omezený na prvních 5 výstupních produktů s preferencí variantních položek.
- `sakypaky` i `sakypaky-test` jsou součástí `POST /internal/rebuild/all`.
- Varianty se skládají podle `ITEMGROUP_ID`.
- Varianty používají dodavatelský `CODE` jako Shoptet kód varianty.
- Parent produkt používá `EXTERNAL_ID` ve tvaru `SAKYPAKY-<ITEMGROUP_ID>`.
- Variantní parametr je pouze `Barva`.
- Pokud má skupina duplicitní hodnotu barvy, hodnota se rozliší kódem varianty.
- Dostupnost:
  - `DELIVERY_DATE=0` -> `Skladem`
  - jiná hodnota -> `Dodání X dnů`

## Kategorie a filtr

Importuje se:

- sedací vaky / sedací pytle -> `SEDACÍ VAKY` (`ID=914`)
- taburety -> `ŽIDLE > TABURETY` (`ID=1155`)
- houpačky -> `ZAHRADNÍ NÁBYTEK > ZAHRADNÍ DOPLŃKY` (`ID=1227`)
- stolky -> `STOLY > ODKLÁDACÍ A PŘÍSTAVNÉ STOLKY` (`ID=1269`)
- sety, náplně a opravné sady -> primárně `SEDACÍ VAKY` (`ID=914`)
- ostatní bezpečně rozpoznané doplňky -> `BYTOVÉ DOPLŇKY` (`ID=1173`)

Vynechává se:

- pelechy / psí produkty
- etikety a jmenovky
- obalové materiály
- položky bez `CODE` nebo `PRODUCTNAME`
- položky bez bezpečně namapované kategorie

## MVP rozsah

- Identifikace produktu:
  - `CODE`
- Bezpečná základní pole:
  - `NAME`
  - `EAN`, pokud je ve zdroji
  - `PRICE_VAT`, pokud je ve zdroji
  - `AVAILABILITY` z `DELIVERY_DATE`
  - `MANUFACTURER` / `SUPPLIER`
  - hlavní obrázek a alternativní obrázky
  - cílová kategorie

## Otevřené otázky

- Po Shoptet test importu doladit případné okrajové mapování kategorií.
- Ověřit, jestli klient chce do ostrého feedu i všechny doplňky typu výhodné sety a opravné sady, nebo jen hlavní nábytek.

## Akceptační kritéria

- `go run ./cmd/rebuild --supplier sakypaky` vytvoří validní XML.
- `go run ./cmd/rebuild --supplier sakypaky-test` vytvoří validní XML s 5 výstupními produkty.
- Výstup je dostupný na `/feeds/sakypaky.xml`.
- Test výstup je dostupný na `/feeds/sakypaky-test.xml`.
- Výstup nepřekročí Shoptet limity.
- Chyba downloadu nebo transformace nepřepíše poslední validní XML.
- `/status` ukazuje výsledek posledního běhu.

## Testy

- Unit test běžné Sakypaky položky.
- Unit test variantní skupiny podle `ITEMGROUP_ID` a parametru `Barva`.
- Unit test duplicitní hodnoty barvy.
- Unit test odřezání částečné barvy z parent názvu.
- Unit test vynechání pet/etiket a neznámých kategorií.
- Unit test mapování kategorií.
- Unit test `sakypaky-test` limitu na 5 výstupních produktů.
- Rebuild test přes fixture-backed downloader.
- Reálný lokální rebuild proti živému zdroji:
  - `go run ./cmd/rebuild --supplier sakypaky-test`
  - `go run ./cmd/rebuild --supplier sakypaky`
- Lokální XML kontrola přes `xmllint --noout`.
- Ruční kontrola přes Shoptet XML validátor.

## Ověření 2026-06-12

- `go run ./cmd/rebuild --supplier sakypaky-test`:
  - `productsRead=1008`
  - `productsEmitted=5`
  - `productsSkipped=6`
  - `itemsWithVariants=5`
  - `variantsEmitted=20`
- `go run ./cmd/rebuild --supplier sakypaky`:
  - `productsRead=1008`
  - `productsEmitted=92`
  - `productsSkipped=6`
  - `itemsWithVariants=75`
  - `variantsEmitted=985`
- Výstupní soubory:
  - `sakypaky-test.xml`: 5 `SHOPITEM`
  - `sakypaky.xml`: 92 `SHOPITEM`
- `xmllint --noout` prošel pro oba výstupy.
