ALTER TABLE CONTFUL_ENT.contful_asset_folders ADD (slug VARCHAR2(255) DEFAULT '');
ALTER TABLE CONTFUL_ENT.contful_asset_folders ADD (path VARCHAR2(500) DEFAULT '');
ALTER TABLE CONTFUL_ENT.contful_asset_folders ADD (sort_order NUMBER DEFAULT 0);
ALTER TABLE CONTFUL_ENT.contful_asset_folders ADD (created_by VARCHAR2(36));
COMMIT;