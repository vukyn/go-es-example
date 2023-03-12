
--- Query count in-stock
-- select wi.sku, count(1) as sku_count from wms_inventory wi 
-- where wi.status_id in (2,3,4,5,6,7,8,9,10,11,14,15,16,17,20,21,22,23,26,27,28,29,30,33)
-- group by sku


--- Query count committed
-- select wi.sku, count(1) as sku_count from wms_inventory wi 
-- where wi.status_id in (7,9,11,12)
-- group by sku
-- order by sku_count desc