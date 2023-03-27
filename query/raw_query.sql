--- Query count in-stock
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (2,3,4,5,6,7,8,9,10,11,14,15,16,17,20,21,22,23,26,27,28,29,30,33)
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count committed
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (7,9,11,12)
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count receving
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (2,26,33)
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count available
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id = 6
and wi.product_status_id = 1
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count test
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id = 6
and wi.product_status_id = 7
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count pick order
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (7,8,9)
and wi.sales_order_type = 'ORDER'
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count pack order
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (10,11)
and wi.sales_order_type = 'ORDER'
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count pick IT
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (20,21)
and wi.sales_order_type = 'INTERNAL_TRANSFER'
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc


--- Query count pack IT
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.status_id in (22,23)
and wi.sales_order_type = 'INTERNAL_TRANSFER'
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc

--- Query count UP
select wi.sku, ww.warehouse_name, count(1) as sku_count from wms_inventory wi 
left join wms_warehouse ww on ww.warehouse_id = wi.warehouse_id
where wi.product_status_id in (5,6,7,8)
group by wi.sku, wi.warehouse_id
order by wi.warehouse_id desc
