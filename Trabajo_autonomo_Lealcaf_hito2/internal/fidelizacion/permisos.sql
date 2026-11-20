REVOKE ALL ON SCHEMA public FROM PUBLIC;
GRANT USAGE ON SCHEMA public TO lealcaf_api;
REVOKE ALL ON ALL TABLES IN SCHEMA public FROM lealcaf_api;
REVOKE ALL ON ALL SEQUENCES IN SCHEMA public FROM lealcaf_api;
GRANT SELECT ON sucursals, clientes, recompensas, cuentas TO lealcaf_api;
GRANT SELECT, INSERT ON compras, movimiento_puntos, canjes TO lealcaf_api;
GRANT UPDATE (estado) ON canjes TO lealcaf_api;
GRANT UPDATE (id) ON clientes TO lealcaf_api;
GRANT USAGE ON SEQUENCE compras_id_seq, movimiento_puntos_id_seq, canjes_id_seq TO lealcaf_api;
