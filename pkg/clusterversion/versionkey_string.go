// Code generated by "stringer -type=VersionKey"; DO NOT EDIT.

package clusterversion

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[Version19_1-0]
	_ = x[Version19_2-1]
	_ = x[VersionContainsEstimatesCounter-2]
	_ = x[VersionSecondaryIndexColumnFamilies-3]
	_ = x[VersionNamespaceTableWithSchemas-4]
	_ = x[VersionProtectedTimestamps-5]
	_ = x[VersionPrimaryKeyChanges-6]
	_ = x[VersionAuthLocalAndTrustRejectMethods-7]
	_ = x[VersionPrimaryKeyColumnsOutOfFamilyZero-8]
	_ = x[VersionNoExplicitForeignKeyIndexIDs-9]
	_ = x[VersionHashShardedIndexes-10]
	_ = x[VersionCreateRolePrivilege-11]
	_ = x[VersionStatementDiagnosticsSystemTables-12]
	_ = x[VersionSchemaChangeJob-13]
	_ = x[VersionSavepoints-14]
	_ = x[Version20_1-15]
	_ = x[VersionStart20_2-16]
	_ = x[VersionGeospatialType-17]
	_ = x[VersionEnums-18]
	_ = x[VersionRangefeedLeases-19]
	_ = x[VersionAlterColumnTypeGeneral-20]
	_ = x[VersionAlterSystemJobsAddCreatedByColumns-21]
	_ = x[VersionAddScheduledJobsTable-22]
	_ = x[VersionUserDefinedSchemas-23]
	_ = x[VersionNoOriginFKIndexes-24]
	_ = x[VersionClientRangeInfosOnBatchResponse-25]
	_ = x[VersionNodeMembershipStatus-26]
	_ = x[VersionRangeStatsRespHasDesc-27]
	_ = x[VersionMinPasswordLength-28]
	_ = x[VersionAbortSpanBytes-29]
	_ = x[VersionAlterSystemJobsAddSqllivenessColumnsAddNewSystemSqllivenessTable-30]
	_ = x[VersionMaterializedViews-31]
	_ = x[VersionBox2DType-32]
	_ = x[VersionLeasedDatabaseDescriptors-33]
	_ = x[VersionUpdateScheduledJobsSchema-34]
	_ = x[VersionCreateLoginPrivilege-35]
	_ = x[VersionHBAForNonTLS-36]
	_ = x[Version20_2-37]
	_ = x[VersionStart21_1-38]
	_ = x[VersionEmptyArraysInInvertedIndexes-39]
}

const _VersionKey_name = "Version19_1Version19_2VersionContainsEstimatesCounterVersionSecondaryIndexColumnFamiliesVersionNamespaceTableWithSchemasVersionProtectedTimestampsVersionPrimaryKeyChangesVersionAuthLocalAndTrustRejectMethodsVersionPrimaryKeyColumnsOutOfFamilyZeroVersionNoExplicitForeignKeyIndexIDsVersionHashShardedIndexesVersionCreateRolePrivilegeVersionStatementDiagnosticsSystemTablesVersionSchemaChangeJobVersionSavepointsVersion20_1VersionStart20_2VersionGeospatialTypeVersionEnumsVersionRangefeedLeasesVersionAlterColumnTypeGeneralVersionAlterSystemJobsAddCreatedByColumnsVersionAddScheduledJobsTableVersionUserDefinedSchemasVersionNoOriginFKIndexesVersionClientRangeInfosOnBatchResponseVersionNodeMembershipStatusVersionRangeStatsRespHasDescVersionMinPasswordLengthVersionAbortSpanBytesVersionAlterSystemJobsAddSqllivenessColumnsAddNewSystemSqllivenessTableVersionMaterializedViewsVersionBox2DTypeVersionLeasedDatabaseDescriptorsVersionUpdateScheduledJobsSchemaVersionCreateLoginPrivilegeVersionHBAForNonTLSVersion20_2VersionStart21_1VersionEmptyArraysInInvertedIndexes"

var _VersionKey_index = [...]uint16{0, 11, 22, 53, 88, 120, 146, 170, 207, 246, 281, 306, 332, 371, 393, 410, 421, 437, 458, 470, 492, 521, 562, 590, 615, 639, 677, 704, 732, 756, 777, 848, 872, 888, 920, 952, 979, 998, 1009, 1025, 1060}

func (i VersionKey) String() string {
	if i < 0 || i >= VersionKey(len(_VersionKey_index)-1) {
		return "VersionKey(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _VersionKey_name[_VersionKey_index[i]:_VersionKey_index[i+1]]
}
