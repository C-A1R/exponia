-- name: ListFilmStocks :many
SELECT
    fs.id,
    fs.name,
    fs.manufacturer,
    fs.iso,
    fct.code AS color_type,
    fs.created_at
FROM film_stocks fs
JOIN film_color_types fct
    ON fct.id = fs.color_type_id
ORDER BY fs.manufacturer, fs.name;


-- name: GetFilmStockByID :one
SELECT
    fs.id,
    fs.name,
    fs.manufacturer,
    fs.iso,
    fct.code AS color_type,
    fs.created_at
FROM film_stocks fs
JOIN film_color_types fct
    ON fct.id = fs.color_type_id
WHERE fs.id = $1;


-- name: ListFilmStockFormats :many
SELECT
    ff.id,
    ff.code,
    ff.name
FROM film_stock_formats fsf
JOIN film_formats ff
    ON ff.id = fsf.format_id
WHERE fsf.film_stock_id = $1
ORDER BY ff.id;

-- name: ListAllFilmStockFormats :many
SELECT
    fsf.film_stock_id,
    ff.id,
    ff.code,
    ff.name
FROM film_stock_formats fsf
JOIN film_formats ff
    ON ff.id = fsf.format_id
ORDER BY fsf.film_stock_id, ff.id;