// Utility function to escape HTML to prevent XSS
function escapeHtml(unsafe) {
  if (typeof unsafe !== "string") return unsafe;
  return unsafe
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#039;");
}

// DOM Elements
const classNameInput = document.getElementById("className");
const groupNameInput = document.getElementById("groupName");
const groupNumberInput = document.getElementById("groupNumber");
const selectedGroupSelect = document.getElementById("selectedGroup");
const fullNameInput = document.getElementById("fullName");
const userNameInput = document.getElementById("userName");
const emailInput = document.getElementById("email");
const passwordLengthInput = document.getElementById("passwordLength");
const passwordErrorP = document.getElementById("passwordError");
const userDataErrorP = document.getElementById("userDataError");
const addGroupBtn = document.getElementById("addGroupBtn");
const addStudentBtn = document.getElementById("addStudentBtn");
const generateJSONBtn = document.getElementById("generateJSONBtn");
const exportBtn = document.getElementById("exportBtn");
const clearJSONBtn = document.getElementById("clearJSONBtn");
const themeToggle = document.getElementById("themeToggle");
const groupsList = document.getElementById("groupsList");
const groupCount = document.getElementById("groupCount");
const body = document.body;

// Export Modal Elements
const exportModal = document.getElementById("exportModal");
const closeModal = document.getElementById("closeModal");
const tabBtns = document.querySelectorAll(".tab-btn");
const tabPanes = document.querySelectorAll(".tab-pane");

// Copy Buttons
const copyJsonBtn = document.getElementById("copyJsonBtn");
const copyTomlBtn = document.getElementById("copyTomlBtn");
const copyYamlBtn = document.getElementById("copyYamlBtn");
const copyMarkdownBtn = document.getElementById("copyMarkdownBtn");
const copyPdfBtn = document.getElementById("copyPdfBtn");
const copyImageBtn = document.getElementById("copyImageBtn");

// Application state
let classData = {
  name: "",
  groups: [],
};

let selectedGroup = null;
let groupNumberCount = 1;

// Check for saved theme preference and apply
function applySavedTheme() {
  const savedTheme = localStorage.getItem("theme");
  if (savedTheme === "dark") {
    body.classList.add("dark-mode");
  } else if (
    !savedTheme &&
    window.matchMedia("(prefers-color-scheme: dark)").matches
  ) {
    // If no saved preference but OS prefers dark mode
    body.classList.add("dark-mode");
    localStorage.setItem("theme", "dark");
  } else {
    body.classList.remove("dark-mode");
    localStorage.setItem("theme", "light");
  }

  // Update button state
  updateThemeButton();
}

// Update theme toggle button state
function updateThemeButton() {
  const isDark = body.classList.contains("dark-mode");
  const sunIcon = themeToggle.querySelector(".sun-icon");
  const moonIcon = themeToggle.querySelector(".moon-icon");

  sunIcon.style.display = isDark ? "block" : "none";
  moonIcon.style.display = isDark ? "none" : "block";
}

// Initialize theme
applySavedTheme();

// Toggle theme function
function toggleTheme() {
  const isDark = body.classList.toggle("dark-mode");
  localStorage.setItem("theme", isDark ? "dark" : "light");
  updateThemeButton();
}

// Listen for system theme changes (only if no theme is set in localStorage)
window
  .matchMedia("(prefers-color-scheme: dark)")
  .addEventListener("change", (e) => {
    if (!localStorage.getItem("theme")) {
      applySavedTheme();
    }
  });

function generatePassword(length) {
  const clampedLength = Math.max(8, Math.min(length, 128));
  const charset =
    "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789!@#$%^&*()";
  return Array.from(crypto.getRandomValues(new Uint8Array(clampedLength)))
    .map((n) => charset[n % charset.length])
    .join("");
}

function updateGroupSelect() {
  // Store current value to restore selection if possible
  const currentValue = selectedGroupSelect.value;

  // Clear all options except the first one
  while (selectedGroupSelect.options.length > 1) {
    selectedGroupSelect.remove(1);
  }

  // Add group options
  classData.groups.forEach((group) => {
    const opt = document.createElement("option");
    opt.value = group.name;
    opt.textContent = group.name;
    selectedGroupSelect.appendChild(opt);
  });

  // Update group count
  updateGroupCount();

  // Restore selection if possible, otherwise select the first group
  if (selectedGroup && classData.groups.includes(selectedGroup)) {
    selectedGroupSelect.value = selectedGroup.name;
  } else if (classData.groups.length > 0) {
    selectedGroup = classData.groups[0];
    selectedGroupSelect.value = selectedGroup.name;
  } else {
    selectedGroupSelect.value = "";
    selectedGroup = null;
  }

  // Update button states
  updateButtonStates();
}

addGroupBtn.addEventListener("click", () => {
  const className = classNameInput.value.trim();
  const groupName = groupNameInput.value.trim();
  let groupNumber = groupNumberInput.value.trim();

  if (!className || !groupName) return;

  if (!classData.name) {
    classData.name = className;
  } else if (classData.name !== className) {
    // If class name changes, clear existing groups or handle as an error
    alert(
      "You can only have one class name per JSON. Please clear the existing data to create a new class."
    );
    return;
  }

  if (!groupNumber) {
    groupNumber = groupNumberCount;
    groupNumberCount++;
  }

  let finalGroupName = `${className}-${groupName}-${groupNumber}`;
  let i = 0;
  while (checkGroupNumberReuse(finalGroupName, classData.groups)) {
    finalGroupName = `${className}-${groupName}-${i}`;
    i++;
  }

  const newGroup = { name: finalGroupName, students: [] };
  classData.groups.push(newGroup);
  selectedGroup = newGroup;
  groupNameInput.value = "";
  groupNumberInput.value = "";
  updateGroupSelect();
  selectedGroupSelect.value = selectedGroup.name;
  renderGroups();
});

selectedGroupSelect.addEventListener("change", () => {
  const groupName = selectedGroupSelect.value;
  selectedGroup = classData.groups.find((group) => group.name === groupName);
});

addStudentBtn.addEventListener("click", () => {
  const re = new RegExp("^[\\w\\-\\.]+@([\\w-]+\\.)+[\\w-]{2,}$", "gm");
  const passwordLength = parseInt(passwordLengthInput.value, 10);

  if (passwordLength > 128) {
    passwordErrorP.textContent =
      "Password length cannot exceed 128 characters.";
    passwordErrorP.style.display = "";
    return;
  } else if (passwordLength < 8) {
    passwordErrorP.textContent =
      "Password length cannot be bellow 8 characters.";
    passwordErrorP.style.display = "";
    return;
  } else {
    passwordErrorP.textContent = "";
    passwordErrorP.style.display = "none";
  }

  if (!selectedGroup) return;

  const fullName = fullNameInput.value.trim();
  const userNameRaw = userNameInput.value.trim();

  if (userNameRaw.length < 3) {
    userDataErrorP.textContent =
      "Username cannot be shorter than 3 characters.";
    userDataErrorP.style.display = "";
    return;
  } else {
    userDataErrorP.textContent = "";
    userDataErrorP.style.display = "none";
  }

  let userName = userNameRaw;
  let i = 0;
  while (checkUserNameReuse(userName, classData.groups)) {
    userName = `${userNameRaw}-${i}`;
    i++;
  }

  const emailRaw = emailInput.value.trim();
  if (!re.test(emailRaw)) {
    userDataErrorP.textContent = "Please enter a valid email address";
    userDataErrorP.style.display = "";
    return;
  } else {
    userDataErrorP.textContent = "";
    userDataErrorP.style.display = "none";
  }

  if (checkEmailReuse(emailRaw, classData.groups)) {
    userDataErrorP.textContent =
      "This email has been reused, please use a different one";
    userDataErrorP.style.display = "";
    return;
  } else {
    userDataErrorP.textContent = "";
    userDataErrorP.style.display = "none";
  }

  if (!fullName || !userName || !emailRaw) return;

  // Add the new student
  selectedGroup.students.push({
    fullName,
    userName,
    password: generatePassword(passwordLength),
    email: emailRaw,
  });

  // Clear the form
  fullNameInput.value = "";
  userNameInput.value = "";
  emailInput.value = "";

  // Update the UI to show the new student
  renderGroups();

  // Make sure the group is expanded and visible
  const groupIndex = classData.groups.indexOf(selectedGroup);
  if (groupIndex !== -1) {
    const groupElement = document.querySelector(
      `[data-group-index="${groupIndex}"]`
    );
    if (groupElement) {
      groupElement.scrollIntoView({ behavior: "smooth", block: "nearest" });
    }
  }
});

exportBtn.addEventListener("click", () => {
  if (!classData.name) return;
  openExportModal();
});

generateJSONBtn.addEventListener("click", async () => {
  if (!classData.name) return;
  const payload = JSON.stringify(classData, null, 2);
  const result = await sendData(payload);
  console.log(result); // Or display the result on the page
});

function updateGroupCount() {
  const count = classData.groups.length;
  groupCount.textContent = `${count} ${count === 1 ? "group" : "groups"}`;
}

function updateButtonStates() {
  const hasGroups = classData.groups.length > 0;
  const hasStudents =
    hasGroups && selectedGroup && selectedGroup.students.length > 0;

  // Enable/disable buttons based on state
  addStudentBtn.disabled = !selectedGroup;
  generateJSONBtn.disabled = !hasGroups || !classData.name;
  exportBtn.disabled = !hasGroups || !classData.name;
  clearJSONBtn.disabled = !hasGroups && !classData.name;
}

// Remove any existing click handlers to prevent duplicates
if (window.groupListClickHandler) {
  groupsList.removeEventListener("click", window.groupListClickHandler);
}

// Define the click handler
window.groupListClickHandler = function (e) {
  // Handle edit group button
  const editGroupBtn = e.target.closest(".edit-group-btn");
  if (editGroupBtn) {
    e.stopPropagation();
    e.preventDefault();
    handleEditGroup(e);
    return;
  }

  // Handle delete group button
  const deleteGroupBtn = e.target.closest(".delete-group-btn");
  if (deleteGroupBtn) {
    e.stopPropagation();
    e.preventDefault();
    handleDeleteGroup(e);
    return;
  }

  // Handle edit student button
  const editStudentBtn = e.target.closest(".edit-student-btn");
  if (editStudentBtn) {
    e.stopPropagation();
    e.preventDefault();
    const existingForm = document.querySelector(".edit-form");
    if (existingForm) {
      const existingStudentItem = existingForm.previousElementSibling;
      if (
        existingStudentItem &&
        existingStudentItem.classList.contains("student-item")
      ) {
        existingStudentItem.style.display = "";
      }
      existingForm.remove();
    }
    handleEditStudent(e);
    return;
  }

  // Handle delete student button
  const deleteStudentBtn = e.target.closest(".delete-student-btn");
  if (deleteStudentBtn) {
    e.stopPropagation();
    e.preventDefault();
    handleDeleteStudent(e);
    return;
  }

  // Handle save group button
  const saveGroupBtn = e.target.closest(".save-group-btn");
  if (saveGroupBtn) {
    e.stopPropagation();
    e.preventDefault();
    const form = e.target.closest(".edit-form");
    const groupCard = e.target.closest(".group-card");
    const groupIndex = parseInt(groupCard.dataset.groupIndex);
    const group = classData.groups[groupIndex];
    const newName = form.querySelector(".edit-group-name")?.value.trim() || "";

    if (newName && newName !== group.name) {
      group.name = newName;
      renderGroups();
      updateGroupSelect();
    } else {
      const groupHeader = groupCard.querySelector(".group-header");
      if (groupHeader) groupHeader.style.display = "";
      form.remove();
    }
    return;
  }

  // Handle cancel group edit button
  const cancelGroupBtn = e.target.closest(".cancel-edit-group");
  if (cancelGroupBtn) {
    e.stopPropagation();
    e.preventDefault();
    const form = e.target.closest(".edit-form");
    const groupCard = e.target.closest(".group-card");
    if (!form || !groupCard) return;

    const groupHeader = groupCard.querySelector(".group-header");
    if (groupHeader) groupHeader.style.display = "";
    form.remove();
    return;
  }
};

// Add the event listener
groupsList.addEventListener("click", window.groupListClickHandler);

function renderGroups() {
  if (!classData.groups.length) {
    groupsList.innerHTML = `
      <div class="empty-state">
        <i class="fas fa-inbox" style="font-size: 3rem; opacity: 0.5;"></i>
        <p>No groups created yet. Add a group to get started!</p>
      </div>`;
    updateGroupCount();
    return;
  }

  groupsList.innerHTML = "";
  classData.groups.forEach((group, groupIndex) => {
    const groupElement = document.createElement("div");
    groupElement.className = "group-card fade-in";
    groupElement.dataset.groupIndex = groupIndex;

    let studentsHtml = "";
    if (group.students && group.students.length > 0) {
      studentsHtml = `
        <div class="student-list">
          ${group.students
            .map(
              (student, studentIndex) => `
            <div class="student-item" data-student-index="${studentIndex}">
              <div class="student-info">
                <div class="student-name">${student.fullName}</div>
                <div class="student-username">@${student.userName}</div>
                <div class="student-email">${student.email}</div>
              </div>
              <div class="student-actions">
                <button class="edit-student-btn btn btn-sm btn-outline" title="Edit student">
                  <i class="fas fa-edit"></i>
                </button>
                <button class="delete-student-btn btn btn-sm btn-error" title="Delete student">
                  <i class="fas fa-trash"></i>
                </button>
              </div>
            </div>
          `
            )
            .join("")}
        </div>
      `;
    } else {
      studentsHtml = `
        <div class="empty-state" style="padding: 1rem;">
          <i class="fas fa-user-graduate" style="font-size: 2rem; opacity: 0.5;"></i>
          <p>No students in this group yet.</p>
        </div>`;
    }

    groupElement.innerHTML = `
      <div class="group-header">
        <h3 class="group-title">
          <i class="fas fa-users mr-2"></i>
          ${group.name}
          <span class="badge">${
            group.students ? group.students.length : 0
          } students</span>
        </h3>
        <div class="group-actions">
          <button class="edit-group-btn btn btn-sm btn-outline" title="Edit group">
            <i class="fas fa-edit"></i>
          </button>
          <button class="delete-group-btn btn btn-sm btn-error" title="Delete group">
            <i class="fas fa-trash"></i>
          </button>
        </div>
      </div>
      ${studentsHtml}
    `;

    groupsList.appendChild(groupElement);
  });

  // Event delegation for dynamic elements
  groupsList.addEventListener("click", (e) => {
    // Handle edit group button
    const editGroupBtn = e.target.closest(".edit-group-btn");
    if (editGroupBtn) {
      e.stopPropagation();
      handleEditGroup(e);
      return;
    }

    // Handle delete group button
    const deleteGroupBtn = e.target.closest(".delete-group-btn");
    if (deleteGroupBtn) {
      e.stopPropagation();
      handleDeleteGroup(e);
      return;
    }

    // Handle edit student button
    const editStudentBtn = e.target.closest(".edit-student-btn");
    if (editStudentBtn) {
      e.stopPropagation();
      // Close any open edit forms first
      const existingForm = document.querySelector(".student-item + .edit-form");
      if (existingForm) {
        const existingStudentItem = existingForm.previousElementSibling;
        if (
          existingStudentItem &&
          existingStudentItem.classList.contains("student-item")
        ) {
          existingStudentItem.style.display = "";
          existingForm.remove();
        }
      }
      handleEditStudent(e);
      return;
    }

    // Handle delete student button
    const deleteStudentBtn = e.target.closest(".delete-student-btn");
    if (deleteStudentBtn) {
      e.stopPropagation();
      handleDeleteStudent(e);
      return;
    }

    // Handle save group button
    const saveGroupBtn = e.target.closest(".save-group-btn");
    if (saveGroupBtn) {
      e.stopPropagation();
      const form = e.target.closest(".edit-form");
      const groupCard = e.target.closest(".group-card");
      const groupIndex = parseInt(groupCard.dataset.groupIndex);
      const group = classData.groups[groupIndex];
      const newName =
        form.querySelector(".edit-group-name")?.value.trim() || "";

      if (newName && newName !== group.name) {
        group.name = newName;
        renderGroups();
        updateGroupSelect();
      } else {
        const groupHeader = groupCard.querySelector(".group-header");
        if (groupHeader) groupHeader.style.display = "";
        form.remove();
      }
      return;
    }

    // Handle cancel group edit button
    const cancelGroupBtn = e.target.closest(".cancel-edit-group");
    if (cancelGroupBtn) {
      e.stopPropagation();
      const form = e.target.closest(".edit-form");
      const groupCard = e.target.closest(".group-card");
      if (!form || !groupCard) return;

      const groupHeader = groupCard.querySelector(".group-header");
      if (groupHeader) groupHeader.style.display = "";
      form.remove();
      return;
    }
  });
}

function handleEditGroup(e) {
  // Close any existing group edit forms first
  const existingForms = document.querySelectorAll(".edit-form");
  existingForms.forEach((form) => form.remove());

  // Show any hidden group headers
  document.querySelectorAll(".group-header").forEach((header) => {
    header.style.display = "";
  });

  const groupCard = e.target.closest(".group-card");
  const groupIndex = parseInt(groupCard.dataset.groupIndex);
  const group = classData.groups[groupIndex];
  const headerElement = groupCard.querySelector(".group-header");

  if (!headerElement) return;

  // Hide the group header while editing
  headerElement.style.display = "none";

  // Create edit form
  const form = document.createElement("div");
  form.className = "edit-form";
  form.innerHTML = `
    <div class="form-group">
      <label class="form-label">Group Name</label>
      <input type="text" class="form-control edit-group-name" value="${group.name}" />
    </div>
    <div class="form-actions">
      <button class="btn btn-outline cancel-edit-group">
        <i class="fas fa-times mr-1"></i>
        Cancel
      </button>
      <button class="btn btn-primary save-group-btn">
        <i class="fas fa-save mr-1"></i>
        Save Changes
      </button>
    </div>
  `;

  // Replace group header with edit form
  const groupHeader = groupCard.querySelector(".group-header");
  groupHeader.style.display = "none";
  groupCard.insertBefore(form, groupHeader.nextSibling);

  // Focus on the input
  form.querySelector("input").focus();

  // Add event listeners
  form.querySelector(".save-group-btn").addEventListener("click", () => {
    const newName = form.querySelector(".edit-group-name").value.trim();
    if (newName && newName !== group.name) {
      group.name = newName;
      renderGroups();
      updateGroupSelect();
    } else {
      groupHeader.style.display = "";
      form.remove();
    }
  });

  form.querySelector(".cancel-edit-group").addEventListener("click", () => {
    groupHeader.style.display = "";
    form.remove();
  });
}

function handleDeleteGroup(e) {
  if (
    !confirm("Are you sure you want to delete this group and all its students?")
  ) {
    return;
  }

  const groupIndex = parseInt(
    e.target.closest(".group-card").dataset.groupIndex
  );
  classData.groups.splice(groupIndex, 1);

  if (selectedGroup && classData.groups.indexOf(selectedGroup) === -1) {
    selectedGroup = null;
  }

  renderGroups();
  updateGroupSelect();
}

function handleEditStudent(e) {
  // Close any open edit forms first
  const existingForm = document.querySelector(".edit-form");
  if (existingForm) {
    const existingStudentItem = existingForm.previousElementSibling;
    if (
      existingStudentItem &&
      existingStudentItem.classList.contains("student-item")
    ) {
      existingStudentItem.style.display = "";
    }
    existingForm.remove();
  }

  const studentItem = e.target.closest(".student-item");
  const groupCard = studentItem.closest(".group-card");
  const groupIndex = parseInt(groupCard.dataset.groupIndex);
  const studentIndex = parseInt(studentItem.dataset.studentIndex);
  const student = classData.groups[groupIndex].students[studentIndex];

  // Create edit form
  const form = document.createElement("div");
  form.className = "edit-form";
  form.innerHTML = `
    <div class="grid grid-cols-1 md:grid-cols-2 gap-4">
      <div class="form-group">
        <label class="form-label">Full Name</label>
        <input type="text" class="form-control edit-student-fullname" value="${escapeHtml(
          student.fullName
        )}" />
      </div>
      <div class="form-group">
        <label class="form-label">Username</label>
        <input type="text" class="form-control edit-student-username" value="${escapeHtml(
          student.userName
        )}" />
      </div>
      <div class="form-group">
        <label class="form-label">Email</label>
        <input type="email" class="form-control edit-student-email" value="${escapeHtml(
          student.email
        )}" />
      </div>
    </div>
    <div class="form-actions">
      <button type="button" class="btn btn-outline cancel-edit-student">
        <i class="fas fa-times mr-1"></i>
        Cancel
      </button>
      <button type="button" class="btn btn-primary save-student-btn">
        <i class="fas fa-save mr-1"></i>
        Save Changes
      </button>
    </div>
  `;

  // Replace student item with edit form
  studentItem.style.display = "none";
  studentItem.parentNode.insertBefore(form, studentItem.nextSibling);

  // Focus on the first input
  const firstInput = form.querySelector("input");
  if (firstInput) firstInput.focus();

  // Add event listeners using event delegation
  // Handle form clicks using event delegation
  form.addEventListener("click", (event) => {
    event.stopPropagation();

    // Save button
    if (event.target.closest(".save-student-btn")) {
      const fullName = form
        .querySelector(".edit-student-fullname")
        .value.trim();
      const userName = form
        .querySelector(".edit-student-username")
        .value.trim();
      const email = form.querySelector(".edit-student-email").value.trim();

      if (fullName && userName && email) {
        student.fullName = fullName;
        student.userName = userName;
        student.email = email;
        renderGroups();
      } else {
        studentItem.style.display = "";
        form.remove();
      }
    }

    // Cancel button
    if (event.target.closest(".cancel-edit-student")) {
      studentItem.style.display = "";
      form.remove();
    }
  });
}

function handleDeleteStudent(e) {
  if (!confirm("Are you sure you want to delete this student?")) {
    return;
  }

  const studentItem = e.target.closest(".student-item");
  const groupIndex = parseInt(
    studentItem.closest(".group-card").dataset.groupIndex
  );
  const studentIndex = parseInt(studentItem.dataset.studentIndex);

  classData.groups[groupIndex].students.splice(studentIndex, 1);
  renderGroups();
}

clearJSONBtn.addEventListener("click", async () => {
  classData = {
    name: "",
    groups: [],
  };
  selectedGroup = null;
  groupNumberCount = 1;
  updateGroupSelect();
  classNameInput.value = "";
  groupNameInput.value = "";
  groupNumberInput.value = "";
  fullNameInput.value = "";
  userNameInput.value = "";
  emailInput.value = "";
  groupNumberCount = 1;
  renderGroups();
  passwordLengthInput.value = 12;
  passwordErrorP.textContent = "";
  passwordErrorP.style.display = "none";
  userDataErrorP.textContent = "";
  userDataErrorP.style.display = "none";
});

// Modal and Tab Functions
function openExportModal() {
  updateExportPreviews();
  exportModal.classList.add("show");
  document.body.style.overflow = "hidden";
}

function closeExportModal() {
  exportModal.classList.remove("show");
  document.body.style.overflow = "";
}

function switchTab(tabName) {
  // Remove active class from all tabs and panes
  tabBtns.forEach((btn) => btn.classList.remove("active"));
  tabPanes.forEach((pane) => pane.classList.remove("active"));

  // Add active class to selected tab and pane
  const activeTabBtn = document.querySelector(`[data-tab="${tabName}"]`);
  const activeTabPane = document.getElementById(tabName);

  if (activeTabBtn && activeTabPane) {
    activeTabBtn.classList.add("active");
    activeTabPane.classList.add("active");
    updateCurrentTab(tabName);
    // Don't call updateExportPreviews here - only call it when data changes
  }
}

// Export Functions
function updateExportPreviews() {
  const jsonPreview = document.getElementById("jsonPreview");
  const tomlPreview = document.getElementById("tomlPreview");
  const yamlPreview = document.getElementById("yamlPreview");
  const markdownPreview = document.getElementById("markdownPreview");

  if (jsonPreview) {
    jsonPreview.textContent = JSON.stringify(classData, null, 2);
    jsonPreview.contentEditable = false;
    jsonPreview.style.cursor = "default";
  }

  if (tomlPreview) {
    tomlPreview.textContent = convertToTOML(classData);
    tomlPreview.contentEditable = false;
    tomlPreview.style.cursor = "default";
  }

  if (yamlPreview) {
    yamlPreview.textContent = convertToYAML(classData);
    yamlPreview.contentEditable = false;
    yamlPreview.style.cursor = "default";
  }

  if (markdownPreview) {
    markdownPreview.textContent = convertToMarkdown(classData);
    markdownPreview.contentEditable = false;
    markdownPreview.style.cursor = "default";
  }
}

function convertToTOML(data) {
  let toml = `[${data.name}]\n`;
  data.groups.forEach((group, index) => {
    toml += `[[${data.name}.groups]]\n`;
    toml += `name = "${group.name}"\n`;
    toml += `students = [\n`;
    group.students.forEach((student) => {
      toml += `  { fullName = "${student.fullName}", userName = "${student.userName}", email = "${student.email}" },\n`;
    });
    toml += `]\n\n`;
  });
  return toml;
}

function convertToYAML(data) {
  let yaml = `${data.name}:\n`;
  yaml += `  groups:\n`;
  data.groups.forEach((group) => {
    yaml += `    - name: "${group.name}"\n`;
    yaml += `      students:\n`;
    group.students.forEach((student) => {
      yaml += `        - fullName: "${student.fullName}"\n`;
      yaml += `          userName: "${student.userName}"\n`;
      yaml += `          email: "${student.email}"\n`;
    });
  });
  return yaml;
}

function convertToMarkdown(data) {
  let markdown = `# ${data.name}\n\n`;
  data.groups.forEach((group) => {
    markdown += `## ${group.name}\n\n`;
    if (group.students.length > 0) {
      markdown += `| Full Name | Username | Email |\n`;
      markdown += `| --- | --- | --- |\n`;
      group.students.forEach((student) => {
        markdown += `| ${student.fullName} | ${student.userName} | ${student.email} |\n`;
      });
      markdown += `\n`;
    } else {
      markdown += `No students in this group.\n\n`;
    }
  });
  return markdown;
}

// Copy Functions
async function copyToClipboard(text) {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch (err) {
    console.error("Failed to copy text: ", err);
    // Fallback for older browsers
    const textArea = document.createElement("textarea");
    textArea.value = text;
    document.body.appendChild(textArea);
    textArea.select();
    try {
      document.execCommand("copy");
      return true;
    } catch (err2) {
      console.error("Fallback copy failed: ", err2);
      return false;
    } finally {
      document.body.removeChild(textArea);
    }
  }
}

async function copyImageToClipboard(canvas) {
  try {
    canvas.toBlob(async (blob) => {
      if (!blob) {
        throw new Error("Failed to create blob from canvas");
      }
      const item = new ClipboardItem({ "image/png": blob });
      await navigator.clipboard.write([item]);
    });
    return true;
  } catch (err) {
    console.error("Failed to copy image: ", err);
    return false;
  }
}

// Helper Functions
function escapeHtml(text) {
  const div = document.createElement("div");
  div.textContent = text;
  return div.innerHTML;
}

function downloadFile(content, filename, mimeType) {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function showCopyFeedback(success, format) {
  // Create or update feedback message
  let feedback = document.getElementById("copy-feedback");
  if (!feedback) {
    feedback = document.createElement("div");
    feedback.id = "copy-feedback";
    feedback.style.cssText = `
      position: fixed;
      top: 20px;
      right: 20px;
      padding: 12px 20px;
      border-radius: 8px;
      font-size: 14px;
      font-weight: 500;
      z-index: 10000;
      transition: all 0.3s ease;
      box-shadow: 0 4px 12px rgba(0,0,0,0.15);
    `;
    document.body.appendChild(feedback);
  }

  if (success) {
    feedback.textContent = `${format} copied to clipboard!`;
    feedback.style.backgroundColor = "var(--color-success, #10b981)";
    feedback.style.color = "white";
  } else {
    feedback.textContent = `Failed to copy ${format}`;
    feedback.style.backgroundColor = "var(--color-error, #ef4444)";
    feedback.style.color = "white";
  }

  // Auto-hide after 3 seconds
  setTimeout(() => {
    if (feedback.parentNode) {
      feedback.style.opacity = "0";
      setTimeout(() => {
        if (feedback.parentNode) {
          feedback.parentNode.removeChild(feedback);
        }
      }, 300);
    }
  }, 3000);
}

function createImageCanvas() {
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  // Set canvas size
  canvas.width = 800;
  canvas.height = 600;

  // Fill background
  ctx.fillStyle = getComputedStyle(document.body).getPropertyValue(
    "--color-surface"
  );
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  // Add text
  ctx.fillStyle = getComputedStyle(document.body).getPropertyValue(
    "--color-text"
  );
  ctx.font = "20px Arial";
  ctx.fillText(`${classData.name} - Export`, 20, 40);

  let yPosition = 80;
  classData.groups.forEach((group) => {
    ctx.font = "16px Arial";
    ctx.fillText(group.name, 20, yPosition);
    yPosition += 30;

    group.students.forEach((student) => {
      ctx.font = "14px Arial";
      ctx.fillText(
        `${student.fullName} (${student.userName}) - ${student.email}`,
        40,
        yPosition
      );
      yPosition += 25;
    });
    yPosition += 20;
  });

  return canvas;
}

// Edited Content Management
let editedContent = {};
let currentTab = "json";

function getEditedContent(format) {
  return editedContent[format];
}

function setEditedContent(format, content) {
  editedContent[format] = content;
}

function getCurrentClassData() {
  // Check if any format has been edited, starting with JSON since it's the most reliable
  for (const format of ['json', 'toml', 'yaml', 'markdown']) {
    const editedContent = getEditedContent(format);
    if (editedContent) {
      // Try to parse the edited content back to classData format
      try {
        if (format === 'json') {
          return JSON.parse(editedContent);
        } else {
          // For non-JSON formats, we need to convert back to classData
          // For now, we'll just use the original classData and rely on the user
          // editing the specific format they want. This is a limitation of the current approach.
          console.warn(`Edited content found in ${format} format, but cannot convert back to classData format`);
          break;
        }
      } catch (error) {
        console.warn(`Failed to parse edited ${format} content:`, error);
        break;
      }
    }
  }

  // If no edited content found or parsing failed, use original classData
  return classData;
}

function updateCurrentTab(tabName) {
  currentTab = tabName;
}

// Event Listeners for Modal
closeModal.addEventListener("click", closeExportModal);

exportModal.addEventListener("click", (e) => {
  if (e.target === exportModal) {
    closeExportModal();
  }
});

// Tab switching
tabBtns.forEach((btn) => {
  btn.addEventListener("click", () => {
    const tabName = btn.dataset.tab;
    switchTab(tabName);
  });
});

// Export button event listeners
exportJsonBtn.addEventListener("click", () => {
  const content = JSON.stringify(classData, null, 2);
  downloadFile(content, `${classData.name}.json`, "application/json");
});

exportTomlBtn.addEventListener("click", () => {
  const content = convertToTOML(classData);
  downloadFile(content, `${classData.name}.toml`, "application/toml");
});

exportYamlBtn.addEventListener("click", () => {
  const content = convertToYAML(classData);
  downloadFile(content, `${classData.name}.yaml`, "application/yaml");
});

exportMarkdownBtn.addEventListener("click", () => {
  const content = convertToMarkdown(classData);
  downloadFile(content, `${classData.name}.md`, "text/markdown");
});

// Copy button event listeners
copyJsonBtn.addEventListener("click", async () => {
  const content = JSON.stringify(classData, null, 2);
  const success = await copyToClipboard(content);
  showCopyFeedback(success, "JSON");
});

copyTomlBtn.addEventListener("click", async () => {
  const content = convertToTOML(classData);
  const success = await copyToClipboard(content);
  showCopyFeedback(success, "TOML");
});

copyYamlBtn.addEventListener("click", async () => {
  const content = convertToYAML(classData);
  const success = await copyToClipboard(content);
  showCopyFeedback(success, "YAML");
});

copyMarkdownBtn.addEventListener("click", async () => {
  const content = convertToMarkdown(classData);
  const success = await copyToClipboard(content);
  showCopyFeedback(success, "Markdown");
});

copyPdfBtn.addEventListener("click", async () => {
  const content = convertToMarkdown(classData);
  const success = await copyToClipboard(content);
  showCopyFeedback(success, "Markdown (for PDF)");
});

copyImageBtn.addEventListener("click", async () => {
  const canvas = createImageCanvas();
  const success = await copyImageToClipboard(canvas);
  showCopyFeedback(success, "Image");
});

exportPdfBtn.addEventListener("click", async () => {
  try {
    showCopyFeedback(true, "Generating PDF...");

    // Build the HTML content as a string
    const htmlContent = `
      <!DOCTYPE html>
      <html>
        <head>
          <meta charset="utf-8">
          <title>${escapeHtml(classData.name)}</title>
          <style>
            body {
              font-family: Arial, sans-serif;
              font-size: 12px;
              margin: 20px;
              color: #000;
            }
            .header {
              text-align: center;
              margin-bottom: 20px;
              border-bottom: 2px solid #333;
              padding-bottom: 10px;
            }
            .group-title {
              font-size: 14px;
              font-weight: bold;
              margin: 15px 0 5px 0;
              color: #333;
            }
            table {
              width: 100%;
              border-collapse: collapse;
              margin-bottom: 15px;
            }
            th, td {
              border: 1px solid #666;
              padding: 4px;
              text-align: left;
              font-size: 10px;
            }
            th {
              background-color: #f0f0f0;
              font-weight: bold;
            }
            .no-students {
              font-style: italic;
              color: #666;
              margin: 10px 0;
            }
          </style>
        </head>
        <body>
          <div class="header">
            <h1>${escapeHtml(classData.name)}</h1>
            <p>Generated: ${new Date().toLocaleString()}</p>
          </div>

          ${classData.groups
            .map(
              (group) => `
            <div class="group-title">${escapeHtml(group.name)} (${
                group.students.length
              } students)</div>
            ${
              group.students.length > 0
                ? `
              <table>
                <thead>
                  <tr>
                    <th>Full Name</th>
                    <th>Username</th>
                    <th>Email</th>
                  </tr>
                </thead>
                <tbody>
                  ${group.students
                    .map(
                      (student) => `
                    <tr>
                      <td>${escapeHtml(student.fullName)}</td>
                      <td>${escapeHtml(student.userName)}</td>
                      <td>${escapeHtml(student.email)}</td>
                    </tr>
                  `
                    )
                    .join("")}
                </tbody>
              </table>
            `
                : '<p class="no-students">No students in this group.</p>'
            }
          `
            )
            .join("")}
        </body>
      </html>
    `;

    // PDF options
    const opt = {
      margin: 0.5,
      filename: `${classData.name}.pdf`,
      image: { type: "jpeg", quality: 0.95 },
      html2canvas: { scale: 2, useCORS: true },
      jsPDF: { unit: "in", format: "letter", orientation: "portrait" },
    };

    // Generate the PDF directly from HTML string
    await html2pdf().set(opt).from(htmlContent).save();

    showCopyFeedback(true, "PDF downloaded successfully!");
  } catch (error) {
    console.error("PDF generation error:", error);
    showCopyFeedback(false, "Failed to generate PDF");
  }
});

exportImageBtn.addEventListener("click", async () => {
  const selectedFormat = document.querySelector(
    'input[name="imageFormat"]:checked'
  ).value;

  // Check if the browser supports the selected format for canvas
  const canvas = document.createElement("canvas");
  const ctx = canvas.getContext("2d");

  // Test which formats are supported by trying to export a tiny canvas
  const supportedFormats = {};

  // Test common formats
  const testFormats = [
    { ext: "png", mime: "image/png" },
    { ext: "jpeg", mime: "image/jpeg" },
    { ext: "webp", mime: "image/webp" },
    { ext: "avif", mime: "image/avif" },
  ];

  for (const format of testFormats) {
    try {
      // Create a tiny test canvas
      canvas.width = 1;
      canvas.height = 1;
      ctx.fillStyle = "red";
      ctx.fillRect(0, 0, 1, 1);

      await new Promise((resolve, reject) => {
        canvas.toBlob((blob) => {
          if (blob && blob.size > 0) {
            supportedFormats[format.ext] = format.mime;
          }
          resolve();
        }, format.mime);
      });
    } catch (e) {
      // Format not supported
      console.log(`Format ${format.ext} not supported`);
    }
  }

  const requestedMime =
    selectedFormat === "jpg" ? "image/jpeg" : `image/${selectedFormat}`;
  const supportedMime = supportedFormats[selectedFormat];

  if (!supportedMime) {
    showCopyFeedback(
      false,
      `${selectedFormat.toUpperCase()} format not supported by this browser`
    );
    return;
  }

  // Create the actual image canvas
  const imageCanvas = createImageCanvas();

  // Export the image
  imageCanvas.toBlob(
    (blob) => {
      if (!blob) {
        showCopyFeedback(false, "Failed to create image");
        return;
      }

      const url = URL.createObjectURL(blob);
      const link = document.createElement("a");
      link.href = url;
      link.download = `${classData.name}.${selectedFormat}`;
      document.body.appendChild(link);
      link.click();
      document.body.removeChild(link);
      URL.revokeObjectURL(url);
    },
    requestedMime,
    0.9
  );
});

function checkUserNameReuse(name, groupsArray) {
  for (const group of groupsArray) {
    for (const student of group.students) {
      if (student.userName === name) {
        return true;
      }
    }
  }
  return false;
}

function checkEmailReuse(email, groupsArray) {
  for (const group of groupsArray) {
    for (const student of group.students) {
      if (student.email === email) {
        return true;
      }
    }
  }
  return false;
}

function checkGroupNumberReuse(groupname, groupsArray) {
  return groupsArray.some((group) => group.name === groupname);
}

document.addEventListener("DOMContentLoaded", () => {
  // Function to set a cookie
  function setCookie(name, value, days) {
    const expires = new Date();
    expires.setTime(expires.getTime() + days * 24 * 60 * 60 * 1000);
    document.cookie = `${name}=${value};expires=${expires.toUTCString()};path=/`;
  }

  // Function to get a cookie value
  function getCookie(name) {
    const nameEQ = name + "=";
    const ca = document.cookie.split(";");
    for (let i = 0; i < ca.length; i++) {
      let c = ca[i];
      while (c.charAt(0) === " ") c = c.substring(1, c.length);
      if (c.indexOf(nameEQ) === 0) return c.substring(nameEQ.length, c.length);
    }
    return null;
  }

  // Apply saved theme from cookie on page load
  const savedTheme = getCookie("theme");
  if (savedTheme === "dark") {
    body.classList.add("dark-mode");
  } else if (savedTheme === "light") {
    body.classList.remove("dark-mode");
  }

  // Update the theme button state
  updateThemeButton();

  // Add theme toggle event listener
  themeToggle.addEventListener("click", toggleTheme);
});
