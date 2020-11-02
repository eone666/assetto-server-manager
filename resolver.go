package acsm

import (
	"net/http"
)

type Resolver struct {
	store           Store
	templateLoader  TemplateLoader
	reloadTemplates bool

	raceManager           *RaceManager
	carManager            *CarManager
	trackManager          *TrackManager
	championshipManager   *ChampionshipManager
	accountManager        *AccountManager
	discordManager        *DiscordManager
	notificationManager   *NotificationManager
	scheduledRacesManager *ScheduledRacesManager
	raceWeekendManager    *RaceWeekendManager
	blockListManager      *BlockListManager

	viewRenderer          *Renderer
	serverProcess         ServerProcess
	raceControl           *RaceControl
	raceControlHub        *RaceControlHub
	contentManagerWrapper *ContentManagerWrapper
	acsrClient            *ACSRClient
	udpPluginAdapter      *UDPPluginAdapter

	// handlers
	baseHandler                 *BaseHandler
	quickRaceHandler            *QuickRaceHandler
	customRaceHandler           *CustomRaceHandler
	championshipsHandler        *ChampionshipsHandler
	accountHandler              *AccountHandler
	auditLogHandler             *AuditLogHandler
	carsHandler                 *CarsHandler
	tracksHandler               *TracksHandler
	weatherHandler              *WeatherHandler
	penaltiesHandler            *PenaltiesHandler
	penaltiesManager            *PenaltiesManager
	resultsHandler              *ResultsHandler
	scheduledRacesHandler       *ScheduledRacesHandler
	contentUploadHandler        *ContentUploadHandler
	raceControlHandler          *RaceControlHandler
	serverAdministrationHandler *ServerAdministrationHandler
	raceWeekendHandler          *RaceWeekendHandler
	customChecksumHandler       *CustomChecksumHandler
	strackerHandler             *StrackerHandler
	healthCheck                 *HealthCheck
	kissMyRankHandler           *KissMyRankHandler
	realPenaltyHandler          *RealPenaltyHandler
}

func NewResolver(templateLoader TemplateLoader, reloadTemplates bool, store Store) (*Resolver, error) {
	r := &Resolver{
		templateLoader:  templateLoader,
		reloadTemplates: reloadTemplates,
		store:           store,
	}

	if err := r.initACSRClient(); err != nil {
		return nil, err
	}

	if err := r.initViewRenderer(); err != nil {
		return nil, err
	}

	return r, nil
}

func (r *Resolver) initViewRenderer() error {
	if r.viewRenderer != nil {
		return nil
	}

	viewRenderer, err := NewRenderer(r.templateLoader, r.store, r.resolveServerProcess(), r.reloadTemplates)

	if err != nil {
		return err
	}

	r.viewRenderer = viewRenderer

	return nil
}

func (r *Resolver) initACSRClient() error {
	serverOptions, err := r.store.LoadServerOptions()

	if err != nil {
		return err
	}

	r.acsrClient = NewACSRClient(serverOptions.ACSRAccountID, serverOptions.ACSRAPIKey, serverOptions.EnableACSR)

	return nil
}

func (r *Resolver) ResolveStore() Store {
	return r.store
}

func (r *Resolver) resolveServerProcess() ServerProcess {
	if r.serverProcess != nil {
		return r.serverProcess
	}

	r.serverProcess = NewAssettoServerProcess(r.ResolveStore(), r.resolveContentManagerWrapper())
	r.serverProcess.SetPlugin(r.resolveUDPPluginAdapter())

	return r.serverProcess
}

func (r *Resolver) resolveUDPPluginAdapter() *UDPPluginAdapter {
	if r.udpPluginAdapter != nil {
		return r.udpPluginAdapter
	}

	r.udpPluginAdapter = NewUDPPluginAdapter(
		r.resolveRaceManager(),
		r.ResolveRaceControl(),
		r.resolveChampionshipManager(),
		r.resolveRaceWeekendManager(),
		r.resolveContentManagerWrapper(),
	)

	return r.udpPluginAdapter
}

func (r *Resolver) resolveContentManagerWrapper() *ContentManagerWrapper {
	if r.contentManagerWrapper != nil {
		return r.contentManagerWrapper
	}

	r.contentManagerWrapper = NewContentManagerWrapper(
		r.ResolveStore(),
		r.resolveCarManager(),
		r.resolveTrackManager(),
	)

	return r.contentManagerWrapper
}

func (r *Resolver) resolveRaceManager() *RaceManager {
	if r.raceManager != nil {
		return r.raceManager
	}

	r.raceManager = NewRaceManager(
		r.store,
		r.resolveServerProcess(),
		r.resolveCarManager(),
		r.resolveTrackManager(),
		r.resolveNotificationManager(),
		r.ResolveRaceControl(),
	)

	return r.raceManager
}

func (r *Resolver) resolveBaseHandler() *BaseHandler {
	if r.baseHandler != nil {
		return r.baseHandler
	}

	r.baseHandler = NewBaseHandler(r.viewRenderer)

	return r.baseHandler
}

func (r *Resolver) resolveCustomRaceHandler() *CustomRaceHandler {
	if r.customRaceHandler != nil {
		return r.customRaceHandler
	}

	r.customRaceHandler = NewCustomRaceHandler(
		r.resolveBaseHandler(),
		r.resolveRaceManager(),
		r.ResolveStore(),
		r.resolveChampionshipManager(),
		r.resolveRaceWeekendManager(),
	)

	return r.customRaceHandler
}

func (r *Resolver) resolveAccountManager() *AccountManager {
	if r.accountManager != nil {
		return r.accountManager
	}

	r.accountManager = NewAccountManager(r.store)

	return r.accountManager
}

func (r *Resolver) resolveAccountHandler() *AccountHandler {
	if r.accountHandler != nil {
		return r.accountHandler
	}

	r.accountHandler = NewAccountHandler(r.resolveBaseHandler(), r.store, r.resolveAccountManager())

	return r.accountHandler
}

func (r *Resolver) resolveQuickRaceHandler() *QuickRaceHandler {
	if r.quickRaceHandler != nil {
		return r.quickRaceHandler
	}

	r.quickRaceHandler = NewQuickRaceHandler(r.resolveBaseHandler(), r.resolveRaceManager())

	return r.quickRaceHandler
}

func (r *Resolver) resolveAuditLogHandler() *AuditLogHandler {
	if r.auditLogHandler != nil {
		return r.auditLogHandler
	}

	r.auditLogHandler = NewAuditLogHandler(r.resolveBaseHandler(), r.store)

	return r.auditLogHandler
}

func (r *Resolver) resolveCarManager() *CarManager {
	if r.carManager != nil {
		return r.carManager
	}

	r.carManager = NewCarManager(
		r.resolveTrackManager(),
		config.Server.ScanContentFolderForChanges,
		config.Server.UseCarNameCache,
	)

	return r.carManager
}

func (r *Resolver) resolveCarsHandler() *CarsHandler {
	if r.carsHandler != nil {
		return r.carsHandler
	}

	r.carsHandler = NewCarsHandler(r.resolveBaseHandler(), r.resolveCarManager())

	return r.carsHandler
}

func (r *Resolver) resolveChampionshipManager() *ChampionshipManager {
	if r.championshipManager != nil {
		return r.championshipManager
	}

	r.championshipManager = NewChampionshipManager(
		r.resolveRaceManager(),
		r.acsrClient,
	)

	return r.championshipManager
}

func (r *Resolver) resolveChampionshipsHandler() *ChampionshipsHandler {
	if r.championshipsHandler != nil {
		return r.championshipsHandler
	}

	r.championshipsHandler = NewChampionshipsHandler(r.resolveBaseHandler(), r.resolveChampionshipManager())

	return r.championshipsHandler
}

func (r *Resolver) resolveTrackManager() *TrackManager {
	if r.trackManager != nil {
		return r.trackManager
	}

	r.trackManager = NewTrackManager()

	return r.trackManager
}

func (r *Resolver) resolveTracksHandler() *TracksHandler {
	if r.tracksHandler != nil {
		return r.tracksHandler
	}

	r.tracksHandler = NewTracksHandler(r.resolveBaseHandler(), r.resolveTrackManager())

	return r.tracksHandler
}

func (r *Resolver) resolveWeatherHandler() *WeatherHandler {
	if r.weatherHandler != nil {
		return r.weatherHandler
	}

	r.weatherHandler = NewWeatherHandler(r.resolveBaseHandler())

	return r.weatherHandler
}

func (r *Resolver) resolvePenaltiesHandler() *PenaltiesHandler {
	if r.penaltiesHandler != nil {
		return r.penaltiesHandler
	}

	r.penaltiesHandler = NewPenaltiesHandler(r.resolveBaseHandler(), r.resolvePenaltiesManager())

	return r.penaltiesHandler
}

func (r *Resolver) resolvePenaltiesManager() *PenaltiesManager {
	if r.penaltiesHandler != nil {
		return r.penaltiesManager
	}

	r.penaltiesManager = NewPenaltiesManager(r.ResolveStore())

	return r.penaltiesManager
}

func (r *Resolver) resolveResultsHandler() *ResultsHandler {
	if r.resultsHandler != nil {
		return r.resultsHandler
	}

	r.resultsHandler = NewResultsHandler(r.resolveBaseHandler(), r.ResolveStore())

	return r.resultsHandler
}

func (r *Resolver) resolveScheduledRacesManager() *ScheduledRacesManager {
	if r.scheduledRacesManager != nil {
		return r.scheduledRacesManager
	}

	r.scheduledRacesManager = NewScheduledRacesManager(r.ResolveStore())

	return r.scheduledRacesManager
}

func (r *Resolver) resolveScheduledRacesHandler() *ScheduledRacesHandler {
	if r.scheduledRacesHandler != nil {
		return r.scheduledRacesHandler
	}

	r.scheduledRacesHandler = NewScheduledRacesHandler(r.resolveBaseHandler(), r.resolveScheduledRacesManager())

	return r.scheduledRacesHandler
}

func (r *Resolver) resolveServerAdministrationHandler() *ServerAdministrationHandler {
	if r.serverAdministrationHandler != nil {
		return r.serverAdministrationHandler
	}

	r.serverAdministrationHandler = NewServerAdministrationHandler(
		r.resolveBaseHandler(),
		r.ResolveStore(),
		r.resolveRaceManager(),
		r.resolveChampionshipManager(),
		r.resolveRaceWeekendManager(),
		r.resolveBlockListManager(),
		r.resolveServerProcess(),
		r.acsrClient,
	)

	return r.serverAdministrationHandler
}

func (r *Resolver) resolveBlockListManager() *BlockListManager {
	if r.blockListManager != nil {
		return r.blockListManager
	}

	r.blockListManager = NewBlockListManager()

	return r.blockListManager
}

func (r *Resolver) resolveContentUploadHandler() *ContentUploadHandler {
	if r.contentUploadHandler != nil {
		return r.contentUploadHandler
	}

	r.contentUploadHandler = NewContentUploadHandler(r.resolveBaseHandler(), r.resolveCarManager(), r.resolveTrackManager())

	return r.contentUploadHandler
}

func (r *Resolver) resolveRaceControlHub() *RaceControlHub {
	if r.raceControlHub != nil {
		return r.raceControlHub
	}

	r.raceControlHub = newRaceControlHub()
	go panicCapture(r.raceControlHub.run)

	return r.raceControlHub
}

func (r *Resolver) ResolveRaceControl() *RaceControl {
	if r.raceControl != nil {
		return r.raceControl
	}

	r.raceControl = NewRaceControl(
		r.resolveRaceControlHub(),
		filesystemTrackData{},
		r.resolveServerProcess(),
		r.ResolveStore(),
		r.resolvePenaltiesManager(),
	)

	return r.raceControl
}

func (r *Resolver) resolveRaceControlHandler() *RaceControlHandler {
	if config.Server.PerformanceMode {
		return nil
	}

	if r.raceControlHandler != nil {
		return r.raceControlHandler
	}

	r.raceControlHandler = NewRaceControlHandler(
		r.resolveBaseHandler(),
		r.ResolveStore(),
		r.resolveRaceManager(),
		r.ResolveRaceControl(),
		r.resolveRaceControlHub(),
		r.resolveServerProcess(),
	)

	return r.raceControlHandler
}

func (r *Resolver) resolveRaceWeekendManager() *RaceWeekendManager {
	if r.raceWeekendManager != nil {
		return r.raceWeekendManager
	}

	r.raceWeekendManager = NewRaceWeekendManager(
		r.resolveRaceManager(),
		r.resolveChampionshipManager(),
		r.ResolveStore(),
		r.resolveServerProcess(),
		r.resolveNotificationManager(),
		r.acsrClient,
		r.resolveCarManager(),
	)

	return r.raceWeekendManager
}

func (r *Resolver) resolveRaceWeekendHandler() *RaceWeekendHandler {
	if r.raceWeekendHandler != nil {
		return r.raceWeekendHandler
	}

	r.raceWeekendHandler = NewRaceWeekendHandler(r.resolveBaseHandler(), r.resolveRaceWeekendManager())

	return r.raceWeekendHandler
}

func (r *Resolver) resolveCustomChecksumHandler() *CustomChecksumHandler {
	if r.customChecksumHandler != nil {
		return r.customChecksumHandler
	}

	r.customChecksumHandler = NewCustomChecksumHandler(r.resolveBaseHandler(), r.ResolveStore())

	return r.customChecksumHandler
}

func (r *Resolver) resolveDiscordManager() *DiscordManager {
	if r.discordManager != nil {
		return r.discordManager
	}

	// if manager errors, it will log the error and return discordManager flagged as disabled, so no need to handle err
	r.discordManager, _ = NewDiscordManager(r.store, r.resolveScheduledRacesManager())

	return r.discordManager
}

func (r *Resolver) resolveNotificationManager() *NotificationManager {
	if r.notificationManager != nil {
		return r.notificationManager
	}

	r.notificationManager = NewNotificationManager(r.resolveDiscordManager(), r.resolveCarManager(), r.store)

	return r.notificationManager
}

func (r *Resolver) resolveStrackerHandler() *StrackerHandler {
	if r.strackerHandler != nil {
		return r.strackerHandler
	}

	r.strackerHandler = NewStrackerHandler(r.resolveBaseHandler(), r.ResolveStore())

	return r.strackerHandler
}

func (r *Resolver) resolveHealthCheck() *HealthCheck {
	if r.healthCheck != nil {
		return r.healthCheck
	}

	r.healthCheck = NewHealthCheck(r.ResolveRaceControl(), r.ResolveStore(), r.resolveServerProcess())

	return r.healthCheck
}

func (r *Resolver) resolveKissMyRankHandler() *KissMyRankHandler {
	if r.kissMyRankHandler != nil {
		return r.kissMyRankHandler
	}

	r.kissMyRankHandler = NewKissMyRankHandler(
		r.resolveBaseHandler(),
		r.ResolveStore(),
	)

	return r.kissMyRankHandler
}

func (r *Resolver) resolveRealPenaltyHandler() *RealPenaltyHandler {
	if r.realPenaltyHandler != nil {
		return r.realPenaltyHandler
	}

	r.realPenaltyHandler = NewRealPenaltyHandler(
		r.resolveBaseHandler(),
		r.ResolveStore(),
	)

	return r.realPenaltyHandler
}

func (r *Resolver) ResolveRouter(fs http.FileSystem) http.Handler {
	return Router(
		fs,
		r.resolveQuickRaceHandler(),
		r.resolveCustomRaceHandler(),
		r.resolveChampionshipsHandler(),
		r.resolveAccountHandler(),
		r.resolveAuditLogHandler(),
		r.resolveCarsHandler(),
		r.resolveTracksHandler(),
		r.resolveWeatherHandler(),
		r.resolvePenaltiesHandler(),
		r.resolveResultsHandler(),
		r.resolveContentUploadHandler(),
		r.resolveServerAdministrationHandler(),
		r.resolveRaceControlHandler(),
		r.resolveScheduledRacesHandler(),
		r.resolveRaceWeekendHandler(),
		r.resolveCustomChecksumHandler(),
		r.resolveStrackerHandler(),
		r.resolveHealthCheck(),
		r.resolveKissMyRankHandler(),
		r.resolveRealPenaltyHandler(),
	)
}

type BaseHandler struct {
	viewRenderer *Renderer
}

func NewBaseHandler(viewRenderer *Renderer) *BaseHandler {
	return &BaseHandler{
		viewRenderer: viewRenderer,
	}
}
