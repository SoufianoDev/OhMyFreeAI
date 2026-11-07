const translations = {
  ar: {
    workspace: "مساحة العمل", dashboard: "لوحة التحكم", models: "النماذج", runtime: "بيئة التشغيل",
    preferences: "التفضيلات", settings: "الإعدادات", foundation: "مرحلة التأسيس", localOnly: "محلي", localConnected: "محلي · متصل",
    overview: "نظرة عامة", welcome: "مساحة عملك للذكاء الاصطناعي", welcomeSub: "إدارة النماذج والمحركات وبيئات التشغيل من مكان واحد.",
    foundationStatus: "نسخة تأسيسية", colabNoticeTitle: "تشغيل Colab المجاني قيد مراجعة التوافق",
    colabNoticeBody: "لا يتم تفعيل الاتصال البعيد قبل التحقق من شروط Google. لا يوجد تجاوز للقيود أو ضمان لتوفر GPU.",
    details: "التفاصيل", systemStatus: "حالة النظام", liveState: "الحالة الحالية", runtimeStatus: "بيئة التشغيل",
    notConnected: "غير متصل", runtimeEmpty: "لم يتم إنشاء جلسة تشغيل.", manage: "إدارة البيئة", gatewayStatus: "البوابة المحلية",
    stopped: "متوقفة", notStarted: "لم يتم تشغيل الخدمة.", modelsStatus: "النماذج", verifiedModels: "نماذج تم التحقق منها",
    browseCatalog: "استعراض الكتالوج", recentActivity: "النشاط الأخير", noEvents: "لا توجد أحداث", nothingYet: "لا يوجد نشاط بعد",
    activityWillAppear: "ستظهر هنا عمليات التنزيل والتشغيل والاستعادة بعد تنفيذها.", quickAccess: "وصول سريع", modelCatalog: "كتالوج النماذج",
    modelCatalogSub: "التوافق والموارد المطلوبة", runtimeSetup: "بيئة التشغيل", runtimeSetupSub: "حالة الاتصال والسياسات",
    settingsSub: "تفضيلات الواجهة الأساسية.", footerNote: "مجاني · محلي أولًا · مفتوح المصدر", library: "المكتبة",
    modelsSub: "الكتالوج يميز بين الدعم المثبت والتوافق غير المتحقق.", compatibilityRule: "لا نعتمد اسم العائلة وحده لإثبات التوافق",
    compatibilityRuleSub: "كل نتيجة مرتبطة بإصدار النموذج وملفه والمحرك والـ backend وموارد التشغيل واختبار استدلال فعلي.",
    catalogEmpty: "لم تتم إضافة نماذج بعد", catalogEmptySub: "ستظهر هنا النماذج المدعومة وغير المدعومة وغير المتحقق منها مع أسباب واضحة ومتطلبات التشغيل الدنيا.",
    infrastructure: "البنية التحتية", runtimeSub: "حالة الموارد والسياسات منفصلة عن حالة الاتصال التقنية.", provider: "الموفر",
    policyReview: "مراجعة السياسة مطلوبة", policyGate: "بوابة التوافق مع شروط الخدمة",
    policyDetails: "تقيّد Google بعض أنماط التحكم عن بُعد وتجاوز واجهة الدفتر وتقديم خدمات الويب على بيئات Colab المُدارة المجانية. لذلك لا يمكن تفعيل مسار خادم استدلال بعيد قبل مراجعة نمط الاستخدام المحدد والتأكد من توافقه.",
    policySource: "المصدر: شروط وأسئلة Colab الرسمية.", resourceSnapshot: "لقطة الموارد", resourceUnavailable: "لا توجد بيانات GPU أو RAM أو مساحة قرص موثقة. لا يتم افتراض نوع GPU أو توفره.",
    appearance: "المظهر", appearanceSub: "اختر المظهر الفاتح أو الداكن.", light: "فاتح", dark: "داكن", language: "اللغة",
    languageSub: "العربية والإنجليزية مع اتجاه تخطيط صحيح."
  },
  en: {
    workspace: "Workspace", dashboard: "Dashboard", models: "Models", runtime: "Runtime", preferences: "Preferences",
    settings: "Settings", foundation: "Foundation", localOnly: "Local", localConnected: "Local · Connected", overview: "OVERVIEW", welcome: "Your AI workspace",
    welcomeSub: "Manage models, engines and runtime environments from one place.", foundationStatus: "Foundation build",
    colabNoticeTitle: "Colab Free is pending policy review", colabNoticeBody: "Remote access stays disabled until Google's terms are reviewed. No policy bypass or GPU availability is assumed.",
    details: "Details", systemStatus: "System status", liveState: "Current state", runtimeStatus: "Runtime", notConnected: "Not connected",
    runtimeEmpty: "No runtime session has been created.", manage: "Manage runtime", gatewayStatus: "Local Gateway", stopped: "Stopped",
    notStarted: "Service has not been started.", modelsStatus: "Models", verifiedModels: "Verified models", browseCatalog: "Browse catalog",
    recentActivity: "Recent activity", noEvents: "No events", nothingYet: "Nothing here yet", activityWillAppear: "Downloads, runtime actions and restore events will appear here.",
    quickAccess: "Quick access", modelCatalog: "Model catalog", modelCatalogSub: "Compatibility and resource requirements", runtimeSetup: "Runtime setup",
    runtimeSetupSub: "Connection and policy status", settingsSub: "Basic interface preferences.", footerNote: "Free · Local-first · Open source",
    library: "LIBRARY", modelsSub: "The catalog distinguishes verified support from unverified compatibility.",
    compatibilityRule: "Family names alone do not establish compatibility", compatibilityRuleSub: "Each result is tied to an exact revision, artifact, engine, backend, runtime resources and real inference test.",
    catalogEmpty: "No models added yet", catalogEmptySub: "Supported, unsupported and unverified models will appear here with clear reasons and minimum runtime requirements.",
    infrastructure: "INFRASTRUCTURE", runtimeSub: "Resource and policy state are separate from technical connectivity.", provider: "Provider",
    policyReview: "Policy review required", policyGate: "Terms-of-service compatibility gate",
    policyDetails: "Google restricts certain remote-control patterns, notebook UI bypasses and web-service offerings on free managed Colab runtimes. A remote inference server flow cannot be enabled until the exact usage pattern is reviewed and confirmed compliant.",
    policySource: "Source: official Colab terms and FAQ.", resourceSnapshot: "Resource snapshot", resourceUnavailable: "No verified GPU, RAM or disk data is available. No accelerator type or availability is assumed.",
    appearance: "Appearance", appearanceSub: "Choose light or dark appearance.", light: "Light", dark: "Dark", language: "Language",
    languageSub: "Arabic and English with correct layout direction."
  }
};

let language = "ar";
let theme = "light";
const root = document.documentElement;

function updateNativeLabel() {
  const label = document.querySelector(".environment-label [data-i18n='localOnly']");
  if (!label) return;
  const key = root.dataset.nativeReady === "true" ? "localConnected" : "localOnly";
  label.textContent = translations[language][key];
}

function applyLanguage(next) {
  language = next;
  root.lang = next;
  root.dir = next === "ar" ? "rtl" : "ltr";
  document.querySelectorAll("[data-i18n]").forEach((node) => {
    const key = node.dataset.i18n;
    if (translations[next][key]) node.textContent = translations[next][key];
  });
  document.querySelectorAll("[data-language-choice]").forEach((button) => {
    const selected = button.dataset.languageChoice === next;
    button.classList.toggle("is-selected", selected);
    button.setAttribute("aria-pressed", String(selected));
  });
  document.getElementById("language-toggle").textContent = next === "ar" ? "EN" : "ع";
  document.querySelector(".sidebar").setAttribute("aria-label", next === "ar" ? "التنقل الرئيسي" : "Main navigation");
  const activePage = document.querySelector(".page.is-visible");
  if (activePage) showPage(activePage.id.replace("page-", ""));
  updateNativeLabel();
}

function applyTheme(next) {
  theme = next;
  root.dataset.theme = next;
  document.querySelectorAll("[data-theme-choice]").forEach((button) => {
    const selected = button.dataset.themeChoice === next;
    button.classList.toggle("is-selected", selected);
    button.setAttribute("aria-pressed", String(selected));
  });
}

function showPage(name) {
  const page = document.getElementById(`page-${name}`);
  if (!page) return;
  document.querySelectorAll(".page").forEach((item) => {
    const active = item === page;
    item.classList.toggle("is-visible", active);
    item.hidden = !active;
  });
  document.querySelectorAll("[data-page]").forEach((button) => {
    const active = button.dataset.page === name;
    button.classList.toggle("is-active", active);
    if (active) button.setAttribute("aria-current", "page");
    else button.removeAttribute("aria-current");
  });
  const title = document.querySelector(`[data-page="${name}"] [data-i18n]`);
  const fallback = translations[language][name] || translations[language].dashboard;
  document.getElementById("page-title").textContent = title?.textContent || fallback;
}

document.querySelectorAll("[data-page]").forEach((button) => button.addEventListener("click", () => showPage(button.dataset.page)));
document.querySelectorAll("[data-open-page]").forEach((button) => button.addEventListener("click", () => showPage(button.dataset.openPage)));
document.getElementById("theme-toggle").addEventListener("click", () => applyTheme(theme === "light" ? "dark" : "light"));
document.getElementById("language-toggle").addEventListener("click", () => applyLanguage(language === "ar" ? "en" : "ar"));
document.querySelectorAll("[data-theme-choice]").forEach((button) => button.addEventListener("click", () => applyTheme(button.dataset.themeChoice)));
document.querySelectorAll("[data-language-choice]").forEach((button) => button.addEventListener("click", () => applyLanguage(button.dataset.languageChoice)));
applyLanguage(language);
applyTheme(theme);
showPage("dashboard");


window.addEventListener("ohmyfreeai:bridge-ready", async () => {
  try {
    const snapshot = await window.OhMyFreeAI?.GetSnapshot?.();
    if (!snapshot) throw new Error("Native application snapshot was empty.");
    root.dataset.nativeReady = "true";
    updateNativeLabel();
  } catch (error) {
    root.dataset.nativeReady = "error";
    console.error("Could not connect to the OhMyFreeAI native core:", error);
  }
});
