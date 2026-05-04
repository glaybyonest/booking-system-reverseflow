package app

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog"

	infraauth "reserveflow/backend/internal/infrastructure/auth"
	"reserveflow/backend/internal/infrastructure/db"
	infrakafka "reserveflow/backend/internal/infrastructure/kafka"
	"reserveflow/backend/internal/infrastructure/logger"
	infraredis "reserveflow/backend/internal/infrastructure/redis"
	authapp "reserveflow/backend/internal/modules/auth/application"
	authrepo "reserveflow/backend/internal/modules/auth/repository"
	authtransport "reserveflow/backend/internal/modules/auth/transport"
	bookingsapp "reserveflow/backend/internal/modules/bookings/application"
	bookingsrepo "reserveflow/backend/internal/modules/bookings/repository"
	bookingstransport "reserveflow/backend/internal/modules/bookings/transport"
	eventsapp "reserveflow/backend/internal/modules/events/application"
	eventsrepo "reserveflow/backend/internal/modules/events/repository"
	eventstransport "reserveflow/backend/internal/modules/events/transport"
	notificationsapp "reserveflow/backend/internal/modules/notifications/application"
	notificationsrepo "reserveflow/backend/internal/modules/notifications/repository"
	notificationstransport "reserveflow/backend/internal/modules/notifications/transport"
	paymentsapp "reserveflow/backend/internal/modules/payments/application"
	paymentsrepo "reserveflow/backend/internal/modules/payments/repository"
	paymentstransport "reserveflow/backend/internal/modules/payments/transport"
	seatsapp "reserveflow/backend/internal/modules/seats/application"
	seatsrepo "reserveflow/backend/internal/modules/seats/repository"
	seatstransport "reserveflow/backend/internal/modules/seats/transport"
	sessionsapp "reserveflow/backend/internal/modules/sessions/application"
	sessionsrepo "reserveflow/backend/internal/modules/sessions/repository"
	sessionstransport "reserveflow/backend/internal/modules/sessions/transport"
)

type Dependencies struct {
	Config Config
	Log    zerolog.Logger
	DB     *pgxpool.Pool
	Redis  *infraredis.Client
	JWT    *infraauth.Service
	Kafka  *infrakafka.Producer

	AuthHandler          *authtransport.Handler
	EventsHandler        *eventstransport.Handler
	SessionsHandler      *sessionstransport.Handler
	SeatsHandler         *seatstransport.Handler
	BookingsHandler      *bookingstransport.Handler
	PaymentsHandler      *paymentstransport.Handler
	NotificationsHandler *notificationstransport.Handler

	BookingsService      *bookingsapp.Service
	NotificationsService *notificationsapp.Service
}

func NewDependencies(ctx context.Context, cfg Config) (*Dependencies, error) {
	log := logger.New(cfg.LogLevel)
	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}
	redisClient := infraredis.New(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	if err := redisClient.Ping(ctx); err != nil {
		log.Warn().Err(err).Msg("redis ping failed; cache features will retry per request")
	}
	jwtSvc := infraauth.NewService(cfg.JWTAccessSecret, cfg.JWTRefreshSecret, cfg.JWTAccessTTL, cfg.JWTRefreshTTL)
	kafkaProducer := infrakafka.NewProducer(cfg.KafkaBrokers)

	authRepo := authrepo.NewPostgresRepository(pool)
	authService := authapp.NewService(authRepo, jwtSvc)

	eventsRepo := eventsrepo.NewPostgresRepository(pool)
	eventsService := eventsapp.NewService(eventsRepo)

	sessionsRepo := sessionsrepo.NewPostgresRepository(pool)
	sessionsService := sessionsapp.NewService(sessionsRepo)

	seatsRepo := seatsrepo.NewPostgresRepository(pool)
	seatsService := seatsapp.NewService(seatsRepo, redisClient, cfg.SeatmapCacheTTL)

	bookingsRepo := bookingsrepo.NewPostgresRepository(pool)
	bookingsService := bookingsapp.NewService(bookingsRepo, redisClient, cfg.HoldTTL, log)

	paymentsRepo := paymentsrepo.NewPostgresRepository(pool)
	paymentsService := paymentsapp.NewService(paymentsRepo, redisClient, log)

	notificationsRepo := notificationsrepo.NewPostgresRepository(pool)
	notificationsService := notificationsapp.NewService(notificationsRepo)

	return &Dependencies{
		Config:               cfg,
		Log:                  log,
		DB:                   pool,
		Redis:                redisClient,
		JWT:                  jwtSvc,
		Kafka:                kafkaProducer,
		AuthHandler:          authtransport.NewHandler(authService),
		EventsHandler:        eventstransport.NewHandler(eventsService),
		SessionsHandler:      sessionstransport.NewHandler(sessionsService),
		SeatsHandler:         seatstransport.NewHandler(seatsService),
		BookingsHandler:      bookingstransport.NewHandler(bookingsService),
		PaymentsHandler:      paymentstransport.NewHandler(paymentsService),
		NotificationsHandler: notificationstransport.NewHandler(notificationsService),
		BookingsService:      bookingsService,
		NotificationsService: notificationsService,
	}, nil
}

func (d *Dependencies) Close() {
	if d.Redis != nil {
		_ = d.Redis.Close()
	}
	if d.DB != nil {
		d.DB.Close()
	}
}
