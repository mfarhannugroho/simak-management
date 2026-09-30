-- Migration: 000003_master_data_and_registered_items.down.sql
DROP TABLE IF EXISTS worker_registered_items;
DELETE FROM item_categories WHERE id BETWEEN '88888888-8888-8888-8888-888888888801' AND '88888888-8888-8888-8888-888888888807';
