import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import AppShell from '@/layouts/AppShell.vue'
import DashboardOverview from '@/pages/DashboardOverview.vue'
import PlaceholderPage from '@/pages/PlaceholderPage.vue'
import LoginPage from '@/pages/auth/LoginPage.vue'

import CompaniesAndDepartmentsPage from '@/pages/master/CompaniesAndDepartmentsPage.vue'
import WorkersPage from '@/pages/master/WorkersPage.vue'
import ItemCategoriesPage from '@/pages/master/ItemCategoriesPage.vue'
import MyRegisteredItemsPage from '@/pages/profile/MyRegisteredItemsPage.vue'

import SimakListPage from '@/pages/simak/SimakListPage.vue'
import SimakFormPage from '@/pages/simak/SimakFormPage.vue'
import SimakDetailPage from '@/pages/simak/SimakDetailPage.vue'

import ApprovalsQueuePage from '@/pages/approvals/ApprovalsQueuePage.vue'
import ApprovalReviewPage from '@/pages/approvals/ApprovalReviewPage.vue'
import ApprovalMatrixPage from '@/pages/admin/ApprovalMatrixPage.vue'
import UserManagementPage from '@/pages/admin/UserManagementPage.vue'

import IssuancePage from '@/pages/officer/IssuancePage.vue'
import SimakPreviewPage from '@/pages/simak/SimakPreviewPage.vue'
import VerifySimakPage from '@/pages/verification/VerifySimakPage.vue'

import GateDashboardPage from '@/pages/gate/GateDashboardPage.vue'
import GateInspectionPage from '@/pages/gate/GateInspectionPage.vue'
import GateHistoryPage from '@/pages/gate/GateHistoryPage.vue'

import DiscrepancyListPage from '@/pages/discrepancy/DiscrepancyListPage.vue'
import DiscrepancyDetailPage from '@/pages/discrepancy/DiscrepancyDetailPage.vue'

import NotificationCenterPage from '@/pages/notification/NotificationCenterPage.vue'
import ReportsPage from '@/pages/reports/ReportsPage.vue'
import AuditLogPage from '@/pages/audit/AuditLogPage.vue'

const routes: RouteRecordRaw[] = [
  {
    path: '/login',
    name: 'Login',
    component: LoginPage,
    meta: { title: 'Masuk SIMAK', requiresAuth: false },
  },
  {
    path: '/verify/:token',
    name: 'PublicVerifySimak',
    component: VerifySimakPage,
    meta: { title: 'Verifikasi Keaslian SIMAK - PT INKA', requiresAuth: false },
  },
  {
    path: '/',
    component: AppShell,
    meta: { requiresAuth: true },
    children: [
      {
        path: '',
        name: 'DashboardOverview',
        component: DashboardOverview,
        meta: { title: 'Dashboard Overview', phase: 'Phase 1 & 2' },
      },
      {
        path: 'profile/my-items',
        name: 'MyRegisteredItems',
        component: MyRegisteredItemsPage,
        meta: {
          title: 'Barang Bawaan Rutin Pekerja',
          phase: 'Phase 3',
          description: 'Peralatan dan perangkat kerja rutin yang dibawa pekerja ke area PT INKA.',
          roles: ['Applicant', 'Admin'],
        },
      },
      {
        path: 'simak',
        name: 'SimakList',
        component: SimakListPage,
        meta: {
          title: 'Permohonan SIMAK (Applicant Flow)',
          phase: 'Phase 4',
          description: 'Pekerja mendata alat kerja/laptop, melampirkan surat tugas, dan mengajukan SIMAK.',
          permission: 'simak.view',
          roles: ['Applicant', 'Admin'],
        },
      },
      {
        path: 'simak/new',
        name: 'SimakNew',
        component: SimakFormPage,
        meta: {
          title: 'Formulir Pengajuan SIMAK Baru',
          phase: 'Phase 4',
          description: 'Input barang bawaan, nomor seri, jumlah, dan tujuan area workshop INKA.',
          permission: 'simak.create',
          roles: ['Applicant', 'Admin'],
        },
      },
      {
        path: 'simak/:id',
        name: 'SimakDetail',
        component: SimakDetailPage,
        meta: {
          title: 'Detail Permohonan SIMAK',
          phase: 'Phase 4',
          description: 'Rincian izin, daftar barang bawaan, dan pelacakan status approval.',
          permission: 'simak.view',
          roles: ['Applicant', 'Approver', 'SIMAK Officer', 'Gate Officer', 'Supervisor', 'Admin'],
        },
      },
      {
        path: 'simak/:id/edit',
        name: 'SimakEdit',
        component: SimakFormPage,
        meta: {
          title: 'Ubah Permohonan SIMAK',
          phase: 'Phase 4',
          description: 'Perbaikan data barang bawaan atau respon atas permintaan revisi Approver.',
          permission: 'simak.update',
          roles: ['Applicant', 'Admin'],
        },
      },
      {
        path: 'simak/:id/preview',
        name: 'SimakPreview',
        component: SimakPreviewPage,
        meta: {
          title: 'Pratinjau Dokumen Resmi SIMAK',
          phase: 'Phase 6',
          description: 'Format dokumen standar PT INKA (Persero) siap cetak dan simpan PDF.',
          permission: 'simak.view',
        },
      },
      {
        path: 'approvals',
        name: 'ApprovalsQueue',
        component: ApprovalsQueuePage,
        meta: {
          title: 'Antrean Persetujuan SIMAK (Approver Flow)',
          phase: 'Phase 5',
          description: 'Pemeriksaan permohonan SIMAK berjenjang dengan aksi Approve, Reject, atau Request Revision.',
          permission: 'simak.approve',
          roles: ['Approver', 'Supervisor', 'Admin'],
        },
      },
      {
        path: 'approvals/:id',
        name: 'ApprovalReview',
        component: ApprovalReviewPage,
        meta: {
          title: 'Tinjauan & Putusan Persetujuan SIMAK',
          phase: 'Phase 5',
          description: 'Detail verifikasi peralatan bawaan dan eksekusi persetujuan bertingkat.',
          permission: 'simak.approve',
          roles: ['Approver', 'Supervisor', 'Admin'],
        },
      },
      {
        path: 'officer/issuance',
        name: 'OfficerIssuance',
        component: IssuancePage,
        meta: {
          title: 'Penerbitan Dokumen Resmi & QR Code',
          phase: 'Phase 6',
          description: 'Penomoran otomatis SIMAK dan cetak dokumen PDF standar resmi PT INKA dengan kode QR pengaman.',
          permission: 'simak.issue',
          roles: ['SIMAK Officer', 'Admin'],
        },
      },
      {
        path: 'gate',
        name: 'GateInspection',
        component: GateDashboardPage,
        meta: {
          title: 'Pemeriksaan Pos Gerbang (Satpam)',
          phase: 'Phase 7',
          description: 'Mobile/Tablet friendly scan QR, checklist fisik (+/- tombol cepat), dan pencatatan hasil verifikasi.',
          permission: 'inspection.perform',
          roles: ['Gate Officer', 'Admin'],
        },
      },
      {
        path: 'gate/inspect',
        name: 'GateInspectQuick',
        component: GateInspectionPage,
        meta: {
          title: 'Pemeriksaan Fisik SIMAK',
          phase: 'Phase 7',
          description: 'Checklist fisik barang bawaan pekerja dengan tombol +/- responsif.',
          permission: 'inspection.perform',
          roles: ['Gate Officer', 'Admin'],
        },
      },
      {
        path: 'gate/inspect/:simakId',
        name: 'GateInspectDetail',
        component: GateInspectionPage,
        meta: {
          title: 'Pemeriksaan Fisik SIMAK',
          phase: 'Phase 7',
          description: 'Checklist fisik barang bawaan pekerja dengan tombol +/- responsif.',
          permission: 'inspection.perform',
          roles: ['Gate Officer', 'Admin'],
        },
      },
      {
        path: 'gate/history',
        name: 'GateHistory',
        component: GateHistoryPage,
        meta: {
          title: 'Riwayat Verifikasi Pos Gerbang',
          phase: 'Phase 7',
          description: 'Log pergerakan alat kerja masuk dan keluar area pabrik PT INKA.',
          permission: 'inspection.view',
          roles: ['Gate Officer', 'Supervisor', 'Admin'],
        },
      },
      {
        path: 'discrepancies',
        name: 'DiscrepancyLog',
        component: DiscrepancyListPage,
        meta: {
          title: 'Discrepancy Log (Temuan Selisih)',
          phase: 'Phase 8',
          description: 'Investigasi dan tindak lanjut barang mismatch, kelebihan alat, atau barang hilang.',
          permission: 'inspection.resolve_discrepancy',
          roles: ['Supervisor', 'SIMAK Officer', 'Admin'],
        },
      },
      {
        path: 'discrepancies/:id',
        name: 'DiscrepancyDetail',
        component: DiscrepancyDetailPage,
        meta: {
          title: 'Detail & Resolusi Kasus Discrepancy',
          phase: 'Phase 8',
          description: 'Rincian komparasi barang, bukti temuan pos jaga, dan formulir putusan resolusi pengawas.',
          permission: 'inspection.resolve_discrepancy',
          roles: ['Supervisor', 'SIMAK Officer', 'Admin'],
        },
      },
      {
        path: 'notifications',
        name: 'NotificationCenter',
        component: NotificationCenterPage,
        meta: {
          title: 'Pusat Notifikasi & Pemberitahuan',
          phase: 'Phase 9',
          description: 'Daftar seluruh pemberitahuan dan status real-time sistem.',
        },
      },
      {
        path: 'audit',
        name: 'AuditTrail',
        component: AuditLogPage,
        meta: {
          title: 'Audit Trail & Compliance Log',
          phase: 'Phase 11',
          description: 'Pencatatan tamper-proof seluruh aksi pengguna dan perubahan status SIMAK.',
          permission: 'audit.view',
          roles: ['Admin', 'Supervisor'],
        },
      },
      {
        path: 'reports',
        name: 'ReportsAnalytics',
        component: ReportsPage,
        meta: {
          title: 'Rekap & Analitik SIMAK',
          phase: 'Phase 10',
          description: 'Ekspor Excel / PDF dan statistik utilisasi alat kerja serta audit integritas pergerakan.',
        },
      },
      {
        path: 'admin/master-data',
        name: 'MasterData',
        component: CompaniesAndDepartmentsPage,
        meta: {
          title: 'Master Data Organisasi & Perusahaan',
          phase: 'Phase 3',
          description: 'Pengelolaan data PT INKA, divisi/departemen, vendor rekanan.',
          permission: 'system.manage',
          roles: ['Admin'],
        },
      },
      {
        path: 'admin/workers',
        name: 'WorkerManagement',
        component: WorkersPage,
        meta: {
          title: 'Data Pekerja & Teknisi',
          phase: 'Phase 3',
          description: 'Daftar profil pekerja internal dan teknisi rekanan.',
          permission: 'system.manage',
          roles: ['Admin'],
        },
      },
      {
        path: 'admin/item-categories',
        name: 'ItemCategories',
        component: ItemCategoriesPage,
        meta: {
          title: 'Kategori Alat & Peralatan',
          phase: 'Phase 3',
          description: 'Klasifikasi barang bawaan dan pedoman peralatan kerja.',
          permission: 'system.manage',
          roles: ['Admin'],
        },
      },
      {
        path: 'admin/users',
        name: 'UserManagement',
        component: UserManagementPage,
        meta: {
          title: 'Manajemen Pengguna & RBAC',
          phase: 'Phase 2',
          description: 'Pengelolaan 6 role pengguna dan hak akses granular sistem.',
          permission: 'user.manage',
          roles: ['Admin'],
        },
      },
      {
        path: 'admin/approval-matrix',
        name: 'ApprovalMatrixConfig',
        component: ApprovalMatrixPage,
        meta: {
          title: 'Konfigurasi Approval Matrix',
          phase: 'Phase 5',
          description: 'Matriks persetujuan berjenjang fleksibel tanpa hardcode kode program.',
          permission: 'system.manage',
          roles: ['Admin'],
        },
      },
    ],
  },
  {
    path: '/:pathMatch(.*)*',
    redirect: '/',
  },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
})

// Navigation Guard for Authentication and RBAC
router.beforeEach((to, _from, next) => {
  const authStore = useAuthStore()

  // Set document title
  const title = (to.meta?.title as string) || 'SIMAK PT INKA'
  document.title = `${title} — PT INKA (Persero)`

  const requiresAuth = to.meta?.requiresAuth !== false
  const token = localStorage.getItem('simak_access_token')

  // Not authenticated & page requires login -> redirect to /login
  if (requiresAuth && !token) {
    return next({
      path: '/login',
      query: { redirect: to.fullPath },
    })
  }

  // Already authenticated & visits /login -> redirect to /
  if (to.path === '/login' && token) {
    return next('/')
  }

  // Permission & Role Checks
  const requiredRoles = to.meta?.roles as string[] | undefined
  const requiredPermission = to.meta?.permission as string | undefined

  if (token && (requiredRoles || requiredPermission)) {
    // If Admin, always pass
    if (authStore.roles.includes('Admin')) {
      return next()
    }

    if (requiredRoles && requiredRoles.length && !authStore.hasAnyRole(requiredRoles)) {
      alert(`Akses Ditolak: Halaman ini hanya untuk peran ${requiredRoles.join(', ')}.`)
      return next('/')
    }

    if (requiredPermission && !authStore.hasPermission(requiredPermission)) {
      alert(`Akses Ditolak: Anda tidak memiliki izin (${requiredPermission}).`)
      return next('/')
    }
  }

  next()
})
