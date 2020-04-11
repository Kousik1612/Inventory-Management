CREATE DATABASE inventory

CREATE SCHEMA inv

-- Table: inv.products

-- DROP TABLE inv.products;

CREATE TABLE inv.products
(
    product_id integer NOT NULL DEFAULT nextval('inv.products_product_id_seq'::regclass),
    product_type character varying(1000) COLLATE pg_catalog."default",
    seller_id character varying COLLATE pg_catalog."default",
    size character varying COLLATE pg_catalog."default",
    brand character varying COLLATE pg_catalog."default",
    metadata json,
    items integer,
    location character varying COLLATE pg_catalog."default",
    location_id character varying COLLATE pg_catalog."default",
    isavailable boolean,
    CONSTRAINT products_pkey PRIMARY KEY (product_id)
)

TABLESPACE pg_default;

ALTER TABLE inv.products
    OWNER to postgres;


-- FUNCTION: inv.order_products(character varying)

-- DROP FUNCTION inv.order_products(character varying);

CREATE OR REPLACE FUNCTION inv.order_products(
	product_ids character varying DEFAULT ''::character varying)
    RETURNS character varying
    LANGUAGE 'plpgsql'

    COST 100
    VOLATILE 
    
AS $BODY$
declare
    result varchar;
    query varchar;
	exist integer;
	items integer;
begin
    query := 'SELECT 1 FROM inv.products where product_id = ANY (''' || '{' || product_ids || '}' || ''':: int[]) and isavailable=true and items > 0';
    execute query into exist;
    if exist = 1 then
    	    query := 'update inv.products set items = items -1 where product_id = ANY (''' || '{' || product_ids || '}' || ''':: int[]) RETURNING items';
			execute query into items;
		if items = 0 then 
			query := 'update inv.products set isAvailable = false where product_id = ANY (''' || '{' || product_ids || '}' || ''':: int[])';
			execute query;
		end if;	
		RETURN 2;
	end if;
RETURN 1;        
end;
$BODY$;

ALTER FUNCTION inv.order_products(character varying)
    OWNER TO postgres;
