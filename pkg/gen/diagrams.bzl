# Generated by genbzl

DIAGRAMS_SRCS = [
    "//docs/generated/sql/bnf:abort.html",
    "//docs/generated/sql/bnf:add_column.html",
    "//docs/generated/sql/bnf:add_constraint.html",
    "//docs/generated/sql/bnf:alter.html",
    "//docs/generated/sql/bnf:alter_backup.html",
    "//docs/generated/sql/bnf:alter_backup_schedule.html",
    "//docs/generated/sql/bnf:alter_changefeed.html",
    "//docs/generated/sql/bnf:alter_column.html",
    "//docs/generated/sql/bnf:alter_database.html",
    "//docs/generated/sql/bnf:alter_database_add_region.html",
    "//docs/generated/sql/bnf:alter_database_add_super_region.html",
    "//docs/generated/sql/bnf:alter_database_alter_super_region.html",
    "//docs/generated/sql/bnf:alter_database_drop_region.html",
    "//docs/generated/sql/bnf:alter_database_drop_secondary_region.html",
    "//docs/generated/sql/bnf:alter_database_drop_super_region.html",
    "//docs/generated/sql/bnf:alter_database_owner.html",
    "//docs/generated/sql/bnf:alter_database_placement.html",
    "//docs/generated/sql/bnf:alter_database_primary_region.html",
    "//docs/generated/sql/bnf:alter_database_set.html",
    "//docs/generated/sql/bnf:alter_database_set_secondary_region.html",
    "//docs/generated/sql/bnf:alter_database_set_zone_config_extension.html",
    "//docs/generated/sql/bnf:alter_database_survival_goal.html",
    "//docs/generated/sql/bnf:alter_database_to_schema.html",
    "//docs/generated/sql/bnf:alter_ddl.html",
    "//docs/generated/sql/bnf:alter_default_privileges.html",
    "//docs/generated/sql/bnf:alter_func.html",
    "//docs/generated/sql/bnf:alter_func_dep_extension.html",
    "//docs/generated/sql/bnf:alter_func_options.html",
    "//docs/generated/sql/bnf:alter_func_owner.html",
    "//docs/generated/sql/bnf:alter_func_rename.html",
    "//docs/generated/sql/bnf:alter_func_set_schema.html",
    "//docs/generated/sql/bnf:alter_index.html",
    "//docs/generated/sql/bnf:alter_index_partition_by.html",
    "//docs/generated/sql/bnf:alter_index_visible.html",
    "//docs/generated/sql/bnf:alter_partition.html",
    "//docs/generated/sql/bnf:alter_primary_key.html",
    "//docs/generated/sql/bnf:alter_proc.html",
    "//docs/generated/sql/bnf:alter_proc_owner.html",
    "//docs/generated/sql/bnf:alter_proc_rename.html",
    "//docs/generated/sql/bnf:alter_proc_set_schema.html",
    "//docs/generated/sql/bnf:alter_range.html",
    "//docs/generated/sql/bnf:alter_range_relocate.html",
    "//docs/generated/sql/bnf:alter_rename_view.html",
    "//docs/generated/sql/bnf:alter_role.html",
    "//docs/generated/sql/bnf:alter_scatter.html",
    "//docs/generated/sql/bnf:alter_scatter_index.html",
    "//docs/generated/sql/bnf:alter_schema.html",
    "//docs/generated/sql/bnf:alter_sequence.html",
    "//docs/generated/sql/bnf:alter_sequence_options.html",
    "//docs/generated/sql/bnf:alter_sequence_owner.html",
    "//docs/generated/sql/bnf:alter_sequence_set_schema.html",
    "//docs/generated/sql/bnf:alter_table.html",
    "//docs/generated/sql/bnf:alter_table_cmds.html",
    "//docs/generated/sql/bnf:alter_table_locality.html",
    "//docs/generated/sql/bnf:alter_table_owner.html",
    "//docs/generated/sql/bnf:alter_table_partition_by.html",
    "//docs/generated/sql/bnf:alter_table_reset_storage_param.html",
    "//docs/generated/sql/bnf:alter_table_set_schema.html",
    "//docs/generated/sql/bnf:alter_table_set_storage_param.html",
    "//docs/generated/sql/bnf:alter_type.html",
    "//docs/generated/sql/bnf:alter_view.html",
    "//docs/generated/sql/bnf:alter_view_owner.html",
    "//docs/generated/sql/bnf:alter_view_set_schema.html",
    "//docs/generated/sql/bnf:alter_zone_database.html",
    "//docs/generated/sql/bnf:alter_zone_index.html",
    "//docs/generated/sql/bnf:alter_zone_partition.html",
    "//docs/generated/sql/bnf:alter_zone_range.html",
    "//docs/generated/sql/bnf:alter_zone_table.html",
    "//docs/generated/sql/bnf:analyze.html",
    "//docs/generated/sql/bnf:backup.html",
    "//docs/generated/sql/bnf:backup_options.html",
    "//docs/generated/sql/bnf:begin.html",
    "//docs/generated/sql/bnf:begin_transaction.html",
    "//docs/generated/sql/bnf:call.html",
    "//docs/generated/sql/bnf:cancel.html",
    "//docs/generated/sql/bnf:cancel_all_jobs.html",
    "//docs/generated/sql/bnf:cancel_job.html",
    "//docs/generated/sql/bnf:cancel_query.html",
    "//docs/generated/sql/bnf:cancel_session.html",
    "//docs/generated/sql/bnf:check_column_level.html",
    "//docs/generated/sql/bnf:check_table_level.html",
    "//docs/generated/sql/bnf:close_cursor.html",
    "//docs/generated/sql/bnf:col_qualification.html",
    "//docs/generated/sql/bnf:column_table_def.html",
    "//docs/generated/sql/bnf:comment.html",
    "//docs/generated/sql/bnf:commit_prepared.html",
    "//docs/generated/sql/bnf:commit_transaction.html",
    "//docs/generated/sql/bnf:copy.html",
    "//docs/generated/sql/bnf:copy_to.html",
    "//docs/generated/sql/bnf:create.html",
    "//docs/generated/sql/bnf:create_as_col_qual_list.html",
    "//docs/generated/sql/bnf:create_as_constraint_def.html",
    "//docs/generated/sql/bnf:create_changefeed.html",
    "//docs/generated/sql/bnf:create_database.html",
    "//docs/generated/sql/bnf:create_ddl.html",
    "//docs/generated/sql/bnf:create_extension.html",
    "//docs/generated/sql/bnf:create_external_connection.html",
    "//docs/generated/sql/bnf:create_func.html",
    "//docs/generated/sql/bnf:create_index.html",
    "//docs/generated/sql/bnf:create_index_with_storage_param.html",
    "//docs/generated/sql/bnf:create_inverted_index.html",
    "//docs/generated/sql/bnf:create_logical_replication_stream.html",
    "//docs/generated/sql/bnf:create_proc.html",
    "//docs/generated/sql/bnf:create_role.html",
    "//docs/generated/sql/bnf:create_schedule.html",
    "//docs/generated/sql/bnf:create_schedule_for_backup.html",
    "//docs/generated/sql/bnf:create_schedule_for_changefeed.html",
    "//docs/generated/sql/bnf:create_schema.html",
    "//docs/generated/sql/bnf:create_sequence.html",
    "//docs/generated/sql/bnf:create_stats.html",
    "//docs/generated/sql/bnf:create_table.html",
    "//docs/generated/sql/bnf:create_table_as.html",
    "//docs/generated/sql/bnf:create_table_with_storage_param.html",
    "//docs/generated/sql/bnf:create_trigger.html",
    "//docs/generated/sql/bnf:create_type.html",
    "//docs/generated/sql/bnf:create_view.html",
    "//docs/generated/sql/bnf:deallocate.html",
    "//docs/generated/sql/bnf:declare_cursor.html",
    "//docs/generated/sql/bnf:default_value_column_level.html",
    "//docs/generated/sql/bnf:delete.html",
    "//docs/generated/sql/bnf:discard.html",
    "//docs/generated/sql/bnf:drop.html",
    "//docs/generated/sql/bnf:drop_column.html",
    "//docs/generated/sql/bnf:drop_constraint.html",
    "//docs/generated/sql/bnf:drop_database.html",
    "//docs/generated/sql/bnf:drop_ddl.html",
    "//docs/generated/sql/bnf:drop_external_connection.html",
    "//docs/generated/sql/bnf:drop_func.html",
    "//docs/generated/sql/bnf:drop_index.html",
    "//docs/generated/sql/bnf:drop_owned_by.html",
    "//docs/generated/sql/bnf:drop_proc.html",
    "//docs/generated/sql/bnf:drop_role.html",
    "//docs/generated/sql/bnf:drop_schedule.html",
    "//docs/generated/sql/bnf:drop_schema.html",
    "//docs/generated/sql/bnf:drop_sequence.html",
    "//docs/generated/sql/bnf:drop_table.html",
    "//docs/generated/sql/bnf:drop_trigger.html",
    "//docs/generated/sql/bnf:drop_type.html",
    "//docs/generated/sql/bnf:drop_view.html",
    "//docs/generated/sql/bnf:execute.html",
    "//docs/generated/sql/bnf:experimental_audit.html",
    "//docs/generated/sql/bnf:explain.html",
    "//docs/generated/sql/bnf:explain_analyze.html",
    "//docs/generated/sql/bnf:explainable.html",
    "//docs/generated/sql/bnf:export.html",
    "//docs/generated/sql/bnf:family_def.html",
    "//docs/generated/sql/bnf:fetch_cursor.html",
    "//docs/generated/sql/bnf:fingerprint_options.html",
    "//docs/generated/sql/bnf:fingerprint_options_list.html",
    "//docs/generated/sql/bnf:for_locking.html",
    "//docs/generated/sql/bnf:foreign_key_column_level.html",
    "//docs/generated/sql/bnf:foreign_key_table_level.html",
    "//docs/generated/sql/bnf:generic_set.html",
    "//docs/generated/sql/bnf:grant.html",
    "//docs/generated/sql/bnf:import_csv.html",
    "//docs/generated/sql/bnf:import_dump.html",
    "//docs/generated/sql/bnf:import_into.html",
    "//docs/generated/sql/bnf:index_def.html",
    "//docs/generated/sql/bnf:insert.html",
    "//docs/generated/sql/bnf:insert_rest.html",
    "//docs/generated/sql/bnf:iso_level.html",
    "//docs/generated/sql/bnf:joined_table.html",
    "//docs/generated/sql/bnf:legacy_begin.html",
    "//docs/generated/sql/bnf:legacy_end.html",
    "//docs/generated/sql/bnf:legacy_transaction.html",
    "//docs/generated/sql/bnf:like_table_option_list.html",
    "//docs/generated/sql/bnf:limit_clause.html",
    "//docs/generated/sql/bnf:move_cursor.html",
    "//docs/generated/sql/bnf:nonpreparable_set.html",
    "//docs/generated/sql/bnf:not_null_column_level.html",
    "//docs/generated/sql/bnf:offset_clause.html",
    "//docs/generated/sql/bnf:on_conflict.html",
    "//docs/generated/sql/bnf:opt_frame_clause.html",
    "//docs/generated/sql/bnf:opt_locality.html",
    "//docs/generated/sql/bnf:opt_persistence_temp_table.html",
    "//docs/generated/sql/bnf:opt_with_show_fingerprints_options.html",
    "//docs/generated/sql/bnf:opt_with_storage_parameter_list.html",
    "//docs/generated/sql/bnf:pause.html",
    "//docs/generated/sql/bnf:pause_all_jobs.html",
    "//docs/generated/sql/bnf:pause_job.html",
    "//docs/generated/sql/bnf:pause_schedule.html",
    "//docs/generated/sql/bnf:preparable.html",
    "//docs/generated/sql/bnf:prepare.html",
    "//docs/generated/sql/bnf:prepare_transaction.html",
    "//docs/generated/sql/bnf:primary_key_column_level.html",
    "//docs/generated/sql/bnf:primary_key_table_level.html",
    "//docs/generated/sql/bnf:reassign_owned_by.html",
    "//docs/generated/sql/bnf:refresh_materialized_views.html",
    "//docs/generated/sql/bnf:release_savepoint.html",
    "//docs/generated/sql/bnf:rename_column.html",
    "//docs/generated/sql/bnf:rename_constraint.html",
    "//docs/generated/sql/bnf:rename_database.html",
    "//docs/generated/sql/bnf:rename_index.html",
    "//docs/generated/sql/bnf:rename_sequence.html",
    "//docs/generated/sql/bnf:rename_table.html",
    "//docs/generated/sql/bnf:reset.html",
    "//docs/generated/sql/bnf:reset_csetting.html",
    "//docs/generated/sql/bnf:reset_session.html",
    "//docs/generated/sql/bnf:restore.html",
    "//docs/generated/sql/bnf:restore_options.html",
    "//docs/generated/sql/bnf:resume.html",
    "//docs/generated/sql/bnf:resume_all_jobs.html",
    "//docs/generated/sql/bnf:resume_job.html",
    "//docs/generated/sql/bnf:resume_schedule.html",
    "//docs/generated/sql/bnf:revoke.html",
    "//docs/generated/sql/bnf:rollback_prepared.html",
    "//docs/generated/sql/bnf:rollback_transaction.html",
    "//docs/generated/sql/bnf:routine_body.html",
    "//docs/generated/sql/bnf:routine_return.html",
    "//docs/generated/sql/bnf:row_source_extension.html",
    "//docs/generated/sql/bnf:savepoint.html",
    "//docs/generated/sql/bnf:scrub.html",
    "//docs/generated/sql/bnf:scrub_database.html",
    "//docs/generated/sql/bnf:scrub_table.html",
    "//docs/generated/sql/bnf:select.html",
    "//docs/generated/sql/bnf:select_clause.html",
    "//docs/generated/sql/bnf:set_cluster_setting.html",
    "//docs/generated/sql/bnf:set_csetting.html",
    "//docs/generated/sql/bnf:set_exprs_internal.html",
    "//docs/generated/sql/bnf:set_local.html",
    "//docs/generated/sql/bnf:set_operation.html",
    "//docs/generated/sql/bnf:set_or_reset_csetting.html",
    "//docs/generated/sql/bnf:set_rest.html",
    "//docs/generated/sql/bnf:set_rest_more.html",
    "//docs/generated/sql/bnf:set_session.html",
    "//docs/generated/sql/bnf:set_transaction.html",
    "//docs/generated/sql/bnf:show_backup.html",
    "//docs/generated/sql/bnf:show_cluster_setting.html",
    "//docs/generated/sql/bnf:show_columns.html",
    "//docs/generated/sql/bnf:show_commit_timestamp.html",
    "//docs/generated/sql/bnf:show_constraints.html",
    "//docs/generated/sql/bnf:show_create.html",
    "//docs/generated/sql/bnf:show_create_external_connections.html",
    "//docs/generated/sql/bnf:show_create_schedules.html",
    "//docs/generated/sql/bnf:show_databases.html",
    "//docs/generated/sql/bnf:show_default_privileges.html",
    "//docs/generated/sql/bnf:show_default_session_variables_for_role.html",
    "//docs/generated/sql/bnf:show_enums.html",
    "//docs/generated/sql/bnf:show_external_connections.html",
    "//docs/generated/sql/bnf:show_full_scans.html",
    "//docs/generated/sql/bnf:show_functions.html",
    "//docs/generated/sql/bnf:show_grants.html",
    "//docs/generated/sql/bnf:show_indexes.html",
    "//docs/generated/sql/bnf:show_jobs.html",
    "//docs/generated/sql/bnf:show_keys.html",
    "//docs/generated/sql/bnf:show_locality.html",
    "//docs/generated/sql/bnf:show_partitions.html",
    "//docs/generated/sql/bnf:show_procedures.html",
    "//docs/generated/sql/bnf:show_range_for_row.html",
    "//docs/generated/sql/bnf:show_ranges.html",
    "//docs/generated/sql/bnf:show_regions.html",
    "//docs/generated/sql/bnf:show_roles.html",
    "//docs/generated/sql/bnf:show_savepoint_status.html",
    "//docs/generated/sql/bnf:show_schedules.html",
    "//docs/generated/sql/bnf:show_schemas.html",
    "//docs/generated/sql/bnf:show_sequences.html",
    "//docs/generated/sql/bnf:show_session.html",
    "//docs/generated/sql/bnf:show_sessions.html",
    "//docs/generated/sql/bnf:show_statements.html",
    "//docs/generated/sql/bnf:show_stats.html",
    "//docs/generated/sql/bnf:show_survival_goal.html",
    "//docs/generated/sql/bnf:show_tables.html",
    "//docs/generated/sql/bnf:show_trace.html",
    "//docs/generated/sql/bnf:show_transactions.html",
    "//docs/generated/sql/bnf:show_transfer.html",
    "//docs/generated/sql/bnf:show_triggers.html",
    "//docs/generated/sql/bnf:show_types.html",
    "//docs/generated/sql/bnf:show_users.html",
    "//docs/generated/sql/bnf:show_var.html",
    "//docs/generated/sql/bnf:show_zone.html",
    "//docs/generated/sql/bnf:simple_select_clause.html",
    "//docs/generated/sql/bnf:sort_clause.html",
    "//docs/generated/sql/bnf:split_index_at.html",
    "//docs/generated/sql/bnf:split_table_at.html",
    "//docs/generated/sql/bnf:stmt.html",
    "//docs/generated/sql/bnf:stmt_block.html",
    "//docs/generated/sql/bnf:stmt_without_legacy_transaction.html",
    "//docs/generated/sql/bnf:table_clause.html",
    "//docs/generated/sql/bnf:table_constraint.html",
    "//docs/generated/sql/bnf:table_ref.html",
    "//docs/generated/sql/bnf:transaction.html",
    "//docs/generated/sql/bnf:truncate.html",
    "//docs/generated/sql/bnf:unique_column_level.html",
    "//docs/generated/sql/bnf:unique_table_level.html",
    "//docs/generated/sql/bnf:unlisten.html",
    "//docs/generated/sql/bnf:unsplit_index_at.html",
    "//docs/generated/sql/bnf:unsplit_table_at.html",
    "//docs/generated/sql/bnf:update.html",
    "//docs/generated/sql/bnf:upsert.html",
    "//docs/generated/sql/bnf:use.html",
    "//docs/generated/sql/bnf:validate_constraint.html",
    "//docs/generated/sql/bnf:values_clause.html",
    "//docs/generated/sql/bnf:window_definition.html",
    "//docs/generated/sql/bnf:with_clause.html",
]
