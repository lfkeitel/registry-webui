package utils

var (
	DatabaseTableNames = []string{
		"settings",
		"user",
		"organization",
		"team",
		"repo",
		"repo_acl",
		"user_team_org",
	}

	// List table columns as slices of strings

	SettingTableCols = []string{
		"id",
		"value",
	}

	UserTableCols = []string{
		"id",
		"name",
		"displayname",
		"password",
		"created",
	}

	OrganizationTableCols = []string{
		"id",
		"namespace",
		"displayname",
	}

	TeamTableCols = []string{
		"id",
		"orgid",
		"name",
	}

	RepoTableCols = []string{
		"id",
		"name",
		"private",
	}

	RepoACLTableCols = []string{
		"id",
		"repoid",
		"teamid",
		"access",
	}

	UserTeamOrgTableCols = []string{
		"userid",
		"teamid",
		"orgid",
	}
)
