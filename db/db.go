package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/Truenya/dota-search-back/condlog"
	"github.com/Truenya/dota-search-back/data"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	log "github.com/sirupsen/logrus"
)

var DB *pgxpool.Pool
var once sync.Once

func Init(dbName string) *pgxpool.Pool {
	once.Do(func() {
		conn, err := newPostgresConn(dbName)
		DB = conn

		condlog.PanicCond("Unable to connect to database: %v", err)
	})

	return DB
}

func newPostgresConn(dbName string) (*pgxpool.Pool, error) {
	conn, err := pgxpool.New(context.Background(), "postgres://true@localhost:5432/"+dbName)

	if err != nil {
		return conn, fmt.Errorf("failed to create pgpool, %w", err)
	}

	return conn, nil
}

func StoreVKConfig(groups ...data.GroupData) {
	batch := &pgx.Batch{}
	for _, n := range groups {
		batch.Queue(queries["storevkms"], n.ID, n.TopicID, n.MaxMessageID)
	}

	br := DB.SendBatch(context.Background(), batch)
	defer condlog.WarnCond("[storevkms] %v", br.Close())

	for _, g := range groups {
		_, err := br.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				log.Debugf("group %d already exists", g.ID)
				continue
			}
		}
	}
}

func StoreM(m ...data.Message) {
	batch := &pgx.Batch{}
	for _, n := range m {
		batch.Queue(queries["storem"], n.Data, n.Link, n.Timestamp, n.Hash, n.ID)
	}

	br := DB.SendBatch(context.Background(), batch)
	defer condlog.WarnCond("[storem] %v", br.Close())

	for _, mes := range m {
		_, err := br.Exec()
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
				log.Warnf("message %s already exists", mes.Data)
				continue
			}

			log.Errorf("Got database error, failed storing message: %v", err)
		}
	}
}

func GetAll[T any](q string) []T {
	rows, err := DB.Query(context.Background(), q)
	condlog.PanicCond("Query failed: %v", err)

	d, err := pgx.CollectRows(rows, pgx.RowToStructByPos[T])
	condlog.PanicCond("Collect failed: %v", err)

	return d
}

func GetAllM() data.Messages {
	d := GetAll[data.Message](queries["getallm"])

	return data.ToMessages(d...)
}

func StoreP(p data.Player) {
	query := queries["storep"]
	log.Debugf("Query: %s", query)

	r, err := DB.Query(context.Background(), query, p.IP, p.Data, p.MMR, p.Link, p.PossiblePos, p.Timestamp)
	if err != nil {
		log.Errorf("Query failed: %v\n", err)
	}

	r.Close()
}

func UpdateP(p data.Player) {
	query := queries["updatep"]
	log.Debugf("Query: %s", query)

	r, err := DB.Query(context.Background(), query, p.Data, p.MMR, p.Link, p.PossiblePos, p.Timestamp, p.IP)
	condlog.ErrCond("Query failed: %v", err)
	r.Close()
}

func DeleteP(ip string) {
	query := queries["deletep"]
	log.Debugf("Query: %s", query)

	r, err := DB.Query(context.Background(), query, ip)
	condlog.ErrCond("Query delete from player where ip "+ip+" failed %v", err)
	r.Close()
}

func GetAllU() data.Players {
	d := GetAll[data.Player](queries["getallu"])

	return data.ToPlayers(d...)
}

func StoreC(c data.Command) {
	query := queries["storec"]
	log.Debugf("Query: %s", query)

	r, err := DB.Query(context.Background(), query, c.IP, c.Data, c.MMR, c.Link, c.PossiblePos, c.Timestamp)
	condlog.ErrCond("Query failed: %v", err)

	r.Close()
}

func UpdateC(c data.Command) {
	query := queries["updatec"]
	log.Debugf("Query: %s", query)

	r, err := DB.Query(context.Background(), query, c.Data, c.MMR, c.Link, c.PossiblePos, c.Timestamp, c.IP)
	condlog.ErrCond("Query failed: %v", err)

	r.Close()
}

func DeleteC(ip string) {
	query := queries["deletec"]
	log.Debugf("Query: %s", query)

	r, err := DB.Query(context.Background(), query, ip)
	if err != nil {
		log.Errorf("Query failed: %v\n", err)
	}

	r.Close()
}

func GetAllC() data.Commands {
	c := GetAll[data.Command](queries["getallc"])

	return data.ToCommands(c...)
}

func GetAllI() []string {
	var ips []string

	q, err := DB.Query(context.Background(), queries["getalli"])
	if err != nil {
		log.Debugf("Query failed: %v\n", err)
		os.Exit(1)
	}

	var ip string
	for q.Next() {
		if q.Scan(&ip) != nil {
			log.Debugf("Failed to scan ip: %s", ip)
			continue
		}

		ips = append(ips, ip)
	}

	return ips
}

func GetVKMiningSettings() []data.GroupData {
	d := GetAll[data.GroupData](queries["getvkms"])

	return d
}

func GetVKCollectingSettings() int {
	var count int
	err := DB.QueryRow(context.Background(), queries["getvkcs"]).Scan(&count)
	condlog.PanicCond("Query failed: %v", err)

	return count
}

func SetVKCollectingSettings(count int) {
	r, err := DB.Query(context.Background(), queries["storevkcs"], count)
	condlog.PanicCond("Query failed: %v", err)

	r.Close()
}
