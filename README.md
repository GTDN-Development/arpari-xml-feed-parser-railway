# ARPARI XML feed parser / transformer

Mezivrstva pro zpracování produktových, skladových a cenových feedů při migraci e-shopu ARPARI z PrestaShopu na Shoptet.

Projekt má připravovat validní Shoptet XML výstupy z původního katalogu a dodavatelských feedů. Konkrétní technologický stack zatím není rozhodnutý. Tento repozitář aktuálně slouží jako technické zadání a prostor pro budoucí implementaci.

Podrobné obecné zadání je v souboru [ZADANI.md](ZADANI.md).

## Kontext

ARPARI migruje z PrestaShopu na Shoptet. První import původního katalogu do Shoptetu se bude připravovat samostatně z původní SQL databáze. Původní katalog je primární zdroj pro názvy produktů, popisy, URL, SEO, původní kategorie, ruční úpravy a existující sortiment.

Dodavatelské feedy se mají napojovat až následně. Nesmí bez kontroly přepisovat citlivá katalogová data ani vytvářet duplicity.

Základní tok:

```text
vstupní feedy / export z původního e-shopu
-> parser / transformer
-> validní Shoptet XML
-> automatické importy v Shoptetu
```

## Cíl nástroje

Nástroj má umět:

- zakládat nové produkty z dodavatelských feedů,
- aktualizovat existující produkty,
- aktualizovat sklad, ceny a dostupnost,
- zachovat původní katalogová data tam, kde jsou důležitá,
- zabránit duplicitám,
- mapovat dodavatelské kategorie na cílové kategorie v Shoptetu,
- převádět varianty a příplatky do nativního Shoptet modelu,
- validovat XML před publikací,
- držet poslední funkční výstupní feed i při chybě nové transformace,
- logovat běhy, chyby a počty zpracovaných položek.

## Zdroje dat

Plánované zdroje pro první fázi:

- původní PrestaShop export produktů, kategorií a vazeb,
- STIMA katalog: `https://www.stima.cz/userfiles/xml/ITTC_SHT_products.xml`,
- STIMA sklad: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock.xml`,
- STIMA sklad + ceny: `https://www.stima.cz/userfiles/xml/ITTC_SHT_stock_price.xml`,
- Autronic dostupnost: `https://autronic.cz/feeds/availability-feed.xml`,
- případně Autronic katalog: `https://autronic.cz/feeds/product-feed.xml`,
- SEGO: `https://segocz.cz/src/Frontend/Files/Feeds/Catalog/zbozi_123456.xml`,
- HON: `https://www.webshop.officepro-brno.cz/import/HONClientFeed/HONClientFeed.xml`.

Zdroje odložené do druhé fáze:

- Sakypaky: `https://www.sakypaky.cz/export/b2b_partners_cs.xml`,
- Dřevočal: `https://www.matrace-drevocal.cz/feed/`.

Poznámka k Autronic dostupnostnímu feedu: `HEAD` request vrací 404, ale běžný `GET` vrací XML. Parser proto musí používat `GET`.

## Výstupní feedy

Veřejné Shoptet XML výstupy pro první fázi budou dostupné například na:

```text
/feeds/stima-products.xml
/feeds/stima-stock.xml
/feeds/stima-stock-price.xml
/feeds/sego.xml
/feeds/hon.xml
/feeds/autronic-availability.xml
```

Výstupy odložené do druhé fáze:

```text
/feeds/sakypaky.xml
/feeds/drevocal.xml
```

Součástí nástroje má být i stavový výstup se stavem posledních běhů, počtem zpracovaných a přeskočených položek a případnými chybami.

## Bezpečnost publikace feedů

Hlavní pravidlo:

```text
Nikdy nepřepsat poslední funkční XML rozbitým feedem.
```

Doporučený proces:

1. stáhnout zdrojový feed,
2. převést data,
3. aplikovat mapování a pravidla,
4. vygenerovat dočasné XML,
5. zvalidovat výstup,
6. až poté nahradit veřejný XML feed.

Pokud transformace nebo validace selže, Shoptet musí dál dostávat poslední validní XML.

## Párování a pravidla aktualizací

Hlavní identifikátor je kód produktu nebo varianty.

Bude potřeba udržovat mapování:

```text
dodavatelský kód -> Shoptet kód / původní kód z PrestaShopu
```

U existujících produktů z původního e-shopu se bez výslovného povolení nemají automaticky přepisovat:

- název,
- popis,
- URL,
- SEO,
- hlavní kategorie,
- ručně upravené obrázky.

Typicky se mají aktualizovat:

- sklad,
- cena,
- dostupnost,
- EAN,
- variantní data,
- vybrané parametry.

## Provoz na Railway

Řešení bude hostované na Railway a má podporovat:

- veřejně dostupné HTTPS URL pro XML feedy,
- plánované spouštění transformací přes cron nebo job,
- ruční rebuild feedů,
- status endpoint nebo jednoduchý status výstup,
- persistentní uložení posledních validních XML,
- logování běhů a chyb v Railway logs,
- konfiguraci přes environment variables a konfigurační soubory.

Výstupní XML se nesmí spoléhat pouze na lokální ephemeral filesystem, pokud by při redeployi mohla zmizet jediná kopie posledního validního feedu.

## Limity k hlídání

- Shoptet limit přibližně 20 000 položek na jeden feed.
- Maximálně 3 variantní parametry na produkt.
- Maximálně 512 variant na produkt.
- Maximálně přibližně 6 příplatkových parametrů na produkt.
- Maximálně přibližně 128 hodnot jednoho příplatkového parametru.

Známé riziko: STIMA produkt `Židle KR18` má ve feedu přibližně 540 variant, což může překročit limit 512 variant na produkt.

## Dokumentace Shoptet

- [Shoptet XML specifikace](https://developers.shoptet.com/shoptet-tools/shoptet-xml-specification/)
- [Automatické importy produktů](https://podpora.shoptet.cz/automaticke-importy-produktu/)
- [Šablony variant](https://podpora.shoptet.cz/sablony-variant/)
- [Příplatkové parametry](https://podpora.shoptet.cz/priplatkove-parametry/)

## Další rozhodnutí

Před implementací je potřeba rozhodnout:

- technologický stack,
- způsob perzistence posledních validních feedů,
- strukturu konfiguračních souborů,
- formát mapování produktů a kategorií,
- validační strategii proti Shoptet XML,
- způsob plánovaného spouštění na Railway,
- rozsah prvního MVP bez feedů Sakypaky a Dřevočal.
