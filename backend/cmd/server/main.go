// @title SIMAK PT INKA (Persero) API
// @version 1.0
// @description Sistem Pencatatan, Penerbitan, Verifikasi & Audit Barang Bawaan Pekerja
// @BasePath /api/v1
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"
	"simak-backend/config"
	_ "simak-backend/docs"
	"simak-backend/internal/approval"
	"simak-backend/internal/audit"
	"simak-backend/internal/auth"
	"simak-backend/internal/company"
	"simak-backend/internal/dashboard"
	"simak-backend/internal/department"
	"simak-backend/internal/discrepancy"
	"simak-backend/internal/document"
	"simak-backend/internal/inspection"
	"simak-backend/internal/item"
	"simak-backend/internal/notification"
	"simak-backend/internal/report"
	"simak-backend/internal/simak"
	"simak-backend/internal/worker"
	"simak-backend/pkg/database"
	"simak-backend/pkg/logger"
	appMiddleware "simak-backend/pkg/middleware"
	redisclient "simak-backend/pkg/redis"
	"simak-backend/pkg/response"
)

func main() {
	// Load .env file if it exists (ignored in production)
	_ = godotenv.Load()

	// Load application configuration
	cfg := config.Load()

	// Initialise structured logger
	logger.Init(cfg.App.Env)

	// Set Gin mode
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to PostgreSQL
	db := database.Connect(cfg)

	// Connect to Redis
	rdb := redisclient.Connect(cfg)
	_ = rdb // will be used in future phases

	// Initialise Gin router
	router := gin.New()

	// Global middleware
	router.Use(appMiddleware.Recovery())
	router.Use(appMiddleware.RequestID())
	router.Use(appMiddleware.SecurityHeadersMiddleware())
	router.Use(logger.GinLogger())

	// CORS configuration supporting localhost, 127.0.0.1 and configured FrontendURL
	corsConfig := cors.Config{
		AllowOriginFunc: func(origin string) bool {
			if origin == "" {
				return true
			}
			if strings.HasPrefix(origin, "http://localhost") || strings.HasPrefix(origin, "http://127.0.0.1") || strings.HasPrefix(origin, "https://localhost") || strings.HasPrefix(origin, "https://127.0.0.1") {
				return true
			}
			return origin == cfg.App.FrontendURL
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))

	// ──────────────────────────────────────────────────────────────────────────
	// Services & Handlers
	// ──────────────────────────────────────────────────────────────────────────
	authService := auth.NewService(db, rdb, cfg)
	authHandler := auth.NewHandler(authService)

	companyHandler := company.NewHandler(db)
	departmentHandler := department.NewHandler(db)
	workerHandler := worker.NewHandler(db)
	itemHandler := item.NewHandler(db)
	simakService := simak.NewService(db)
	simakHandler := simak.NewHandler(simakService, db)
	approvalService := approval.NewService(db)
	approvalHandler := approval.NewHandler(approvalService, db)
	simakService.SetWorkflowGenerator(approvalService)
	documentService := document.NewService(db, cfg.App.FrontendURL)
	documentHandler := document.NewHandler(documentService, db)
	inspectionService := inspection.NewService(db)
	inspectionHandler := inspection.NewHandler(inspectionService, db)
	discrepancyService := discrepancy.NewService(db)
	discrepancyHandler := discrepancy.NewHandler(discrepancyService, db)

	notificationHub := notification.NewHub(cfg.App.FrontendURL)
	notificationService := notification.NewService(db, notificationHub)
	notificationHandler := notification.NewHandler(notificationService, cfg)

	dashboardService := dashboard.NewService(db)
	dashboardHandler := dashboard.NewHandler(dashboardService, db)
	reportService := report.NewService(db)
	reportHandler := report.NewHandler(reportService)

	auditService := audit.NewService(db)
	auditHandler := audit.NewHandler(auditService, db)

	// Start background cron scheduler for expiring/expired SIMAKs
	notificationService.StartCronScheduler(1 * time.Minute)

	// Wire mismatch hook: automatically generate discrepancy cases and trigger real-time notification
	inspectionService.SetMismatchHook(func(tx *gorm.DB, inspectionID uuid.UUID, simakID uuid.UUID, inspectorID uuid.UUID, items []inspection.InspectionItemInput) error {
		err := discrepancyService.AutoCreateFromInspectionInput(tx, inspectionID, simakID, inspectorID, items)
		if err == nil {
			_ = notificationService.SendToRole("Supervisor", notification.CreateNotificationInput{
				Title:         "Temuan Ketidaksesuaian di Pos Gerbang",
				Message:       "Ditemukan selisih fisik atau barang tidak terdaftar pada SIMAK. Mohon lakukan investigasi di Discrepancy Log.",
				Type:          notification.TypeDiscrepancyCreated,
				ReferenceID:   simakID.String(),
				ReferenceType: "simak",
				ActionURL:     "/discrepancies",
			})
		}
		return err
	})

	// ──────────────────────────────────────────────────────────────────────────
	// Routes
	// ──────────────────────────────────────────────────────────────────────────

	v1 := router.Group("/api/v1")
	{
		// Health check — used by Docker, load balancers, and the frontend to verify
		// the API is reachable and all dependencies are healthy.
		v1.GET("/health", func(c *gin.Context) {
			// Check database connectivity
			dbStatus := "ok"
			if err := database.Ping(db); err != nil {
				dbStatus = "error: " + err.Error()
			}

			// Check Redis connectivity
			redisStatus := "ok"
			if err := redisclient.Ping(rdb); err != nil {
				redisStatus = "error: " + err.Error()
			}

			overallHealthy := dbStatus == "ok" && redisStatus == "ok"

			statusCode := http.StatusOK
			if !overallHealthy {
				statusCode = http.StatusServiceUnavailable
			}

			c.JSON(statusCode, gin.H{
				"success": overallHealthy,
				"message": "SIMAK Management System API",
				"data": gin.H{
					"status":    map[bool]string{true: "healthy", false: "unhealthy"}[overallHealthy],
					"version":   "1.0.0",
					"timestamp": time.Now().Format(time.RFC3339),
					"services": gin.H{
						"database": dbStatus,
						"redis":    redisStatus,
					},
				},
			})
		})

		// Auth routes
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/login", appMiddleware.RateLimiter(rdb, 30, 1*time.Minute), authHandler.Login)
			authGroup.POST("/refresh", authHandler.RefreshToken)

			// Authenticated auth routes
			protectedAuth := authGroup.Group("")
			protectedAuth.Use(appMiddleware.AuthMiddleware(cfg, rdb))
			{
				protectedAuth.POST("/logout", authHandler.Logout)
				protectedAuth.GET("/me", authHandler.GetMe)
				protectedAuth.PUT("/change-password", authHandler.ChangePassword)
			}
		}

		// Company routes
		companyGroup := v1.Group("/companies")
		{
			companyGroup.GET("", companyHandler.ListCompanies)
			companyGroup.GET("/:id", companyHandler.GetCompany)

			protectedComp := companyGroup.Group("")
			protectedComp.Use(appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin"))
			{
				protectedComp.POST("", companyHandler.CreateCompany)
				protectedComp.PUT("/:id", companyHandler.UpdateCompany)
				protectedComp.DELETE("/:id", companyHandler.DeleteCompany)
			}
		}

		// Department routes
		deptGroup := v1.Group("/departments")
		{
			deptGroup.GET("", departmentHandler.ListDepartments)
			deptGroup.GET("/:id", departmentHandler.GetDepartment)

			protectedDept := deptGroup.Group("")
			protectedDept.Use(appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin"))
			{
				protectedDept.POST("", departmentHandler.CreateDepartment)
				protectedDept.PUT("/:id", departmentHandler.UpdateDepartment)
				protectedDept.DELETE("/:id", departmentHandler.DeleteDepartment)
			}
		}

		// Worker routes
		workerGroup := v1.Group("/workers")
		{
			workerGroup.GET("", workerHandler.ListWorkers)
			workerGroup.GET("/:id", workerHandler.GetWorker)

			// Worker items (registered items)
			workerGroup.GET("/:id/registered-items", appMiddleware.AuthMiddleware(cfg, rdb), itemHandler.ListWorkerItems)
			workerGroup.POST("/:id/registered-items", appMiddleware.AuthMiddleware(cfg, rdb), itemHandler.CreateWorkerItem)

			protectedWorker := workerGroup.Group("")
			protectedWorker.Use(appMiddleware.AuthMiddleware(cfg, rdb))
			{
				protectedWorker.POST("", appMiddleware.RequireRole("Admin"), workerHandler.CreateWorker)
				protectedWorker.PUT("/:id", workerHandler.UpdateWorker)
				protectedWorker.DELETE("/:id", appMiddleware.RequireRole("Admin"), workerHandler.DeleteWorker)
			}
		}

		// Item Category routes
		categoryGroup := v1.Group("/item-categories")
		{
			categoryGroup.GET("", itemHandler.ListCategories)

			protectedCat := categoryGroup.Group("")
			protectedCat.Use(appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin"))
			{
				protectedCat.POST("", itemHandler.CreateCategory)
				protectedCat.PUT("/:id", itemHandler.UpdateCategory)
				protectedCat.DELETE("/:id", itemHandler.DeleteCategory)
			}
		}

		// Registered Item direct update/delete routes
		regItemGroup := v1.Group("/registered-items")
		regItemGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			regItemGroup.PUT("/:itemId", itemHandler.UpdateWorkerItem)
			regItemGroup.DELETE("/:itemId", itemHandler.DeleteWorkerItem)
		}

		// SIMAK routes (Permohonan Surat Ijin Membawa Alat Kerja)
		simakGroup := v1.Group("/simak")
		simakGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			simakGroup.GET("", simakHandler.List)
			simakGroup.POST("", simakHandler.Create)
			simakGroup.GET("/:id", simakHandler.GetByID)
			simakGroup.PUT("/:id", simakHandler.Update)
			simakGroup.DELETE("/:id", simakHandler.Delete)
			simakGroup.POST("/:id/submit", simakHandler.Submit)
			simakGroup.POST("/:id/cancel", simakHandler.Cancel)
			simakGroup.POST("/:id/attachments", simakHandler.AddAttachment)
			simakGroup.DELETE("/attachments/:attachmentId", simakHandler.DeleteAttachment)

			// Approver direct actions
			simakGroup.POST("/:id/approve", approvalHandler.Approve)
			simakGroup.POST("/:id/reject", approvalHandler.Reject)
			simakGroup.POST("/:id/revision", approvalHandler.RequestRevision)

			// Document Issuance & retrieval
			simakGroup.POST("/:id/issue", documentHandler.IssueSimak)
			simakGroup.GET("/:id/document", documentHandler.GetDocument)
		}

		// Public QR Code Verification endpoint per PRD §16 (No auth required)
		v1.GET("/verification/:qrToken", documentHandler.PublicVerifyToken)

		// Gate inspection submission
		v1.POST("/verification/inspect", appMiddleware.AuthMiddleware(cfg, rdb), inspectionHandler.RecordInspection)

		// Inspection history & details (Gate Officer, Supervisor, Admin)
		inspectionGroup := v1.Group("/inspections")
		inspectionGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			inspectionGroup.GET("", inspectionHandler.ListInspections)
			inspectionGroup.GET("/:id", inspectionHandler.GetInspection)
		}

		// Gate operational metrics
		v1.GET("/gate/metrics", appMiddleware.AuthMiddleware(cfg, rdb), inspectionHandler.GetMetrics)

		// Realtime WebSocket notification stream
		v1.GET("/ws", notificationHandler.HandleWebSocket)

		// Notification routes (In-App notifications)
		notifGroup := v1.Group("/notifications")
		notifGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			notifGroup.GET("", notificationHandler.ListNotifications)
			notifGroup.GET("/unread-count", notificationHandler.GetUnreadCount)
			notifGroup.PUT("/:id/read", notificationHandler.MarkAsRead)
			notifGroup.PUT("/read-all", notificationHandler.MarkAllAsRead)
		}

		// Discrepancy routes (Supervisor, SIMAK Officer, Admin)
		discrepancyGroup := v1.Group("/discrepancies")
		discrepancyGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			discrepancyGroup.GET("", discrepancyHandler.ListDiscrepancies)
			discrepancyGroup.GET("/metrics", discrepancyHandler.GetMetrics)
			discrepancyGroup.GET("/:id", discrepancyHandler.GetDiscrepancy)
			discrepancyGroup.POST("", discrepancyHandler.CreateDiscrepancy)
			discrepancyGroup.PUT("/:id/review", discrepancyHandler.StartReview)
			discrepancyGroup.PUT("/:id/resolve", discrepancyHandler.ResolveDiscrepancy)
		}

		// Officer queue routes (SIMAK Officer & Admin)
		officerGroup := v1.Group("/officer")
		officerGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			officerGroup.GET("/issuance", documentHandler.GetIssuanceQueue)
		}

		// Approvals workflow queue & detail
		approvalGroup := v1.Group("/approvals")
		approvalGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			approvalGroup.GET("/pending", approvalHandler.GetPendingQueue)
			approvalGroup.GET("/detail/:id", approvalHandler.GetApprovalDetail)
			approvalGroup.POST("/:id/decision", approvalHandler.ProcessDecision)
		}

		// Admin Approval Matrix routes
		matrixGroup := v1.Group("/admin/approval-matrices")
		matrixGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin"))
		{
			matrixGroup.GET("", approvalHandler.ListMatrices)
			matrixGroup.POST("", approvalHandler.CreateMatrix)
			matrixGroup.GET("/:id", approvalHandler.GetMatrix)
			matrixGroup.PUT("/:id", approvalHandler.UpdateMatrix)
			matrixGroup.DELETE("/:id", approvalHandler.DeleteMatrix)
		}

		// Admin User & RBAC Management routes
		adminUserGroup := v1.Group("/admin/users")
		adminUserGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin"))
		{
			adminUserGroup.GET("", authHandler.ListUsers)
			adminUserGroup.PUT("/:id/role", authHandler.UpdateUserRole)
			adminUserGroup.PUT("/:id/status", authHandler.ToggleUserStatus)
		}
		v1.GET("/admin/roles", appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin"), authHandler.ListRoles)

		// Dashboard role-aware summary, charts, and activities
		dashboardGroup := v1.Group("/dashboard")
		dashboardGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			dashboardGroup.GET("/summary", dashboardHandler.GetSummary)
			dashboardGroup.GET("/charts/volume-trend", dashboardHandler.GetVolumeTrend)
			dashboardGroup.GET("/activities", dashboardHandler.GetActivities)
		}

		// Reports & Data Export
		reportGroup := v1.Group("/reports")
		reportGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb))
		{
			reportGroup.GET("/simak", reportHandler.GetSimakReport)
			reportGroup.GET("/verification", reportHandler.GetVerificationReport)
			reportGroup.GET("/discrepancies", reportHandler.GetDiscrepancyReport)
			reportGroup.GET("/export", reportHandler.ExportCSV)
		}

		// Audit Trail & Compliance Log
		auditGroup := v1.Group("/audit-logs")
		auditGroup.Use(appMiddleware.AuthMiddleware(cfg, rdb), appMiddleware.RequireRole("Admin", "Supervisor"))
		{
			auditGroup.GET("", auditHandler.ListAuditLogs)
			auditGroup.GET("/export", auditHandler.ExportCSV)
			auditGroup.GET("/:id", auditHandler.GetByID)
		}
	}

	// Swagger API Documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 404 handler
	router.NoRoute(func(c *gin.Context) {
		response.NotFound(c, "The requested resource was not found")
	})

	// ──────────────────────────────────────────────────────────────────────────
	// Graceful server start & shutdown
	// ──────────────────────────────────────────────────────────────────────────
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine so it doesn't block the graceful shutdown handler.
	go func() {
		log.Printf("🚀 SIMAK API server started on port %s (env: %s)", cfg.App.Port, cfg.App.Env)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server gracefully...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited cleanly.")
}
