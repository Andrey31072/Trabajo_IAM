# Matriz de propiedad de datos

> Estado: 🟡 Revisión | Última actualización: 2026-06-17
> Naming: nombres de entidades en inglés (HALT-DB-NAMING)

Define qué microservicio es el único propietario (escritor) de cada entidad de dominio.
Solo el servicio propietario puede mutar los datos; otros servicios los consumen vía API o eventos.

## Regla

> Si dos servicios necesitan escribir en la misma entidad, el bounded context está mal definido.
> Resolver con un evento o con una entidad local de proyección.

## Matriz

| Entidad | Servicio propietario | BD | Módulo de origen |
|---------|---------------------|----|-----------------|
| `User` | `iam-service` | `iam_db` | M1 |
| `Module` | `iam-service` | `iam_db` | M1 |
| `Feature` | `iam-service` | `iam_db` | M1 |
| `Role` | `iam-service` | `iam_db` | M1 |
| `RoleFeature` | `iam-service` | `iam_db` | M1 |
| `UserRole` | `iam-service` | `iam_db` | M1 |
| `UserScopeOverride` | `iam-service` | `iam_db` | M1 |
| `RefreshToken` | `iam-service` | `iam_db` | M1 |
| `PasswordResetRequest` | `iam-service` | `iam_db` | M1 |
| `AuditLogin` | `iam-service` | `iam_db` | M1 |
| `Macroregion` | `reference-data-service` | `ref_db` | M2 |
| `Microregion` | `reference-data-service` | `ref_db` | M2 |
| `Department` | `reference-data-service` | `ref_db` | M2 |
| `Municipality` | `reference-data-service` | `ref_db` | M2 |
| `TrainingCenter` | `reference-data-service` | `ref_db` | M2 |
| `InstitutionalUnit` | `reference-data-service` | `ref_db` | M2 |
| `Catalog` | `reference-data-service` | `ref_db` | M4 |
| `CatalogDetail` | `reference-data-service` | `ref_db` | M4 |
| `Parameter` | `reference-data-service` | `ref_db` | M4 |
| `TechLine` | `academic-management-service` | `academic_db` | M5 |
| `TechNetwork` | `academic-management-service` | `academic_db` | M5 |
| `KnowledgeNetwork` | `academic-management-service` | `academic_db` | M5 |
| `TrainingProgram` | `academic-management-service` | `academic_db` | M5 |
| `Competency` | `academic-management-service` | `academic_db` | M5 |
| `LearningOutcome` | `academic-management-service` | `academic_db` | M5 |
| `EnrollmentFicha` | `academic-management-service` | `academic_db` | M6 |
| `Environment` | `training-environment-service` | `env_db` | M3 |
| `EnvironmentType` | `training-environment-service` | `env_db` | M3 |
| `InventoryItem` | `training-environment-service` | `env_db` | M3 |
| `Maintenance` | `training-environment-service` | `env_db` | M3 |
| `Reservation` | `training-environment-service` | `env_db` | M3 |
| `AvailabilityRule` | `training-environment-service` | `env_db` | M3 |
| `Schedule` | `scheduling-service` | `scheduling_db` | M8 |
| `ClassSession` | `scheduling-service` | `scheduling_db` | M8 |
| `TimeSlot` | `scheduling-service` | `scheduling_db` | M8 |
| `InstructorScheduleAssignment` | `scheduling-service` | `scheduling_db` | M8 |
| `SchedulingConflict` | `scheduling-service` | `scheduling_db` | M8 |
| `Instructor` | `actors-service` | `actors_db` | M7 |
| `Learner` | `actors-service` | `actors_db` | M7 |
| `Company` | `actors-service` | `actors_db` | M7 |
| `InstructorContract` | `actors-service` | `actors_db` | M7 |
| `CompetencyAssignment` | `actors-service` | `actors_db` | M7 |
| `InstructorAvailabilityException` | `actors-service` | `actors_db` | M7 |
| `ProductiveStage` | `actors-service` | `actors_db` | M7 |
| `CompanyVisit` | `actors-service` | `actors_db` | M7 |
| `ActorImprovementPlan` | `actors-service` | `actors_db` | M7 |
| `ActivityLog` | `actors-service` | `actors_db` | M7 |
| `Document` | `document-service` | `document_db` | transversal |
| `DocumentVersion` | `document-service` | `document_db` | transversal |
| `DocumentTemplate` | `document-service` | `document_db` | transversal |
| `AlertType` | `monitoring-service` | `monitoring_db` | M9 |
| `RiskLevel` | `monitoring-service` | `monitoring_db` | M9 |
| `KpiStatus` | `monitoring-service` | `monitoring_db` | M9 |
| `FichaTracking` | `monitoring-service` | `monitoring_db` | M9 |
| `KpiTracking` | `monitoring-service` | `monitoring_db` | M9 |
| `TrackingSession` | `monitoring-service` | `monitoring_db` | M9 |
| `GeneratedAlert` | `monitoring-service` | `monitoring_db` | M9 |
| `ImprovementPlan` | `monitoring-service` | `monitoring_db` | M9 |
| `SentNotification` | `monitoring-service` | `monitoring_db` | M9 |
| `AuditRecord` | `audit-service` | `audit_db` | transversal |
