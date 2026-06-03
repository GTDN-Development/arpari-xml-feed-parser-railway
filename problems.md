# Problems

Evidence problémů z importů a kontrol dodavatelských feedů.

## STIMA feed

**Datum kontroly:** 2026-06-01

Problém:

Ve STIMA feedu jsou produkty bez fotky. Pokud fotka není ve feedu, parser ji nemá odkud korektně doplnit.

Zjištění:

- `349` produktů je bez fotky a se skladem `0`.
- `87` produktů je bez fotky, ale se skladem `> 0`.

Řešení:

Produkty bez fotky a s explicitním skladem `0` neimportovat. Produkty bez fotky se skladem `> 0` zatím v importu ponechat.

Další zjištění:

- Některé obrázky ve STIMA feedu odkazují na URL, které vrací `404`.
- Některé variantové produkty mají duplicitní kombinace variantních parametrů, takže je Shoptet neumí rozlišit.
- Tyto případy nebudeme uměle opravovat v parseru; jde o problém původních dat od dodavatele.

## HON feed

**Datum kontroly:** 2026-06-01

Produkt: `DY20060001-000119` / `CANTO WHITE PODHLAVNIK - cerna`

Problém:

Shoptet založí produkt, ale nepřidá obrázek. Log hlásí `Invalid image content`.

Detail:

Obrázek ve feedu je, ale jeho URL nevrací JPEG:

```text
https://www.webshop.officepro-brno.cz/import/HONClientFeed/DY20060001-000119_B01_CANTO_WHITE_PODHLAVNIK_CERNA.jpg
```

Závěr:

Chyba je na straně HON/OfficePro URL. Parser ji do Shoptetu propisuje správně.

Řešení:

Oprava URL u dodavatele, případně mirror obrázků.

## Autronic feed

**Datum kontroly:** 2026-06-01

Problém:

Shoptet u části obrázků hlásí timeout při stahování z `autronicshop-cdn.havit.cloud`.

Příčina:

Autronic feed posílá velké galerie a Shoptet má krátký timeout na stažení obrázku.

Závěr:

XML je v pořádku, problém je rychlost/dostupnost externí CDN při importu.

Řešení:

Omezili jsme Autronic feed na max. 10 obrázků na produkt, aby se snížil počet timeoutů.
