package cmd

import (
	"strings"

	appConf "mcop/config"
	"mcop/database/migrations"
	"mcop/internal/config"
	"mcop/internal/database"
	"mcop/internal/log"

	"github.com/spf13/cobra"
	"github.com/uptrace/bun/migrate"
)

func getSvc() *database.DatabaseService {
	confMod := config.New(&appConf.App)
	conf := confMod.Svc.Config()
	log.New(config.Conf[log.Option](confMod.Svc))

	db := database.New(conf.Database.Sql)
	return db.Svc
}

// Migrate Command
func Migrate() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "db",
		Args: NotReqArgs,
	}
	cmd.AddCommand(initCMD())
	cmd.AddCommand(createSQL())
	cmd.AddCommand(createGO())
	cmd.AddCommand(migrateCMD())
	cmd.AddCommand(rollbackCMD()) // Open for use in local only, Don't Commit to PROD
	cmd.AddCommand(statusCMD())
	cmd.AddCommand(markAppliedCMD())
	return cmd
}

func initCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "init",
		Long: "create migration tables",
		Run: func(cmd *cobra.Command, _ []string) {
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			migrator.Init(cmd.Context())
		},
	}
	return cmd
}

func createSQL() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create_sql",
		Long: "create up and down SQL migrations",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.Default()
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			name := strings.Join(args, "_")
			files, err := migrator.CreateSQLMigrations(cmd.Context(), name)
			if err != nil {
				log.Errf("%s", err.Error())
				return
			}
			for _, mf := range files {
				log.Debugf("created migration %s (%s)\n", mf.Name, mf.Path)
			}
		},
	}
	return cmd
}

func createGO() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create_go",
		Long: "create up and down GO migrations",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.Default()
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			name := strings.Join(args, "_")
			file, err := migrator.CreateGoMigration(cmd.Context(), name)
			if err != nil {
				log.Errf("%s", err.Error())
				return
			}
			log.Debugf("created migration %s (%s)\n", file.Name, file.Path)
		},
	}
	return cmd
}

func migrateCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "migrate",
		Run: func(cmd *cobra.Command, args []string) {
			log := log.Default()
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			if err := migrator.Lock(cmd.Context()); err != nil {
				log.Errf("%s", err.Error())
				return
			}
			defer migrator.Unlock(cmd.Context()) //nolint:errcheck

			group, err := migrator.Migrate(cmd.Context())
			if err != nil {
				log.Errf("%s", err.Error())
				return
			}
			if group.IsZero() {
				log.Debugf("there are no new migrations to run (database is up to date)\n")
				return
			}
			log.Debugf("migrated to %s\n", group)
			return
		},
	}
	return cmd
}

func rollbackCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use: "rollback",
		Run: func(cmd *cobra.Command, _ []string) {
			log := log.Default()
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			group, err := migrator.Rollback(cmd.Context())
			if err != nil {
				log.Errf("%s", err.Error())
				return
			}

			if group.ID == 0 {
				log.Debugf("there are no groups to roll back\n")
				return
			}

			log.Debugf("rolled back %s\n", group)
			return
		},
	}
	return cmd
}

func statusCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "status",
		Long: "print migrations status",
		Run: func(cmd *cobra.Command, _ []string) {
			log := log.Default()
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			ms, err := migrator.MigrationsWithStatus(cmd.Context())
			if err != nil {
				log.Errf("%s", err.Error())
				return
			}
			log.Debugf("migrations: %s\n", ms)
			log.Debugf("unapplied migrations: %s\n", ms.Unapplied())
			log.Debugf("last migration group: %s\n", ms.LastGroup())
			return
		},
	}
	return cmd
}

func markAppliedCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "mark_applied",
		Long: "mark migrations as applied without actually running them",
		Run: func(cmd *cobra.Command, _ []string) {
			log := log.Default()
			migrator := migrate.NewMigrator(getSvc().DB(), migrations.Migrations)
			group, err := migrator.Migrate(cmd.Context(), migrate.WithNopMigration())
			if err != nil {
				log.Errf("%s", err.Error())
				return
			}
			if group.IsZero() {
				log.Debugf("there are no new migrations to mark as applied\n")
				return
			}
			log.Debugf("marked as applied %s\n", group)
			return
		},
	}
	return cmd
}
