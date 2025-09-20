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

// Export Buttons
const exportJsonBtn = document.getElementById("exportJsonBtn");
const exportTomlBtn = document.getElementById("exportTomlBtn");
const exportYamlBtn = document.getElementById("exportYamlBtn");
const exportMarkdownBtn = document.getElementById("exportMarkdownBtn");
const exportPdfBtn = document.getElementById("exportPdfBtn");
const exportImageBtn = document.getElementById("exportImageBtn");

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
window.groupListClickHandler = function(e) {
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
    const existingForm = document.querySelector('.edit-form');
    if (existingForm) {
      const existingStudentItem = existingForm.previousElementSibling;
      if (existingStudentItem && existingStudentItem.classList.contains('student-item')) {
        existingStudentItem.style.display = '';
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
    const newName = form.querySelector(".edit-group-name")?.value.trim() || '';

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
  const existingForms = document.querySelectorAll('.edit-form');
  existingForms.forEach(form => form.remove());
  
  // Show any hidden group headers
  document.querySelectorAll('.group-header').forEach(header => {
    header.style.display = '';
  });

  const groupCard = e.target.closest(".group-card");
  const groupIndex = parseInt(groupCard.dataset.groupIndex);
  const group = classData.groups[groupIndex];
  const headerElement = groupCard.querySelector('.group-header');
  
  if (!headerElement) return;
  
  // Hide the group header while editing
  headerElement.style.display = 'none';

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
  exportModal.classList.add('show');
  document.body.style.overflow = 'hidden';
}

function closeExportModal() {
  exportModal.classList.remove('show');
  document.body.style.overflow = '';
}

function switchTab(tabName) {
  // Remove active class from all tabs and panes
  tabBtns.forEach(btn => btn.classList.remove('active'));
  tabPanes.forEach(pane => pane.classList.remove('active'));

  // Add active class to selected tab and pane
  const activeTabBtn = document.querySelector(`[data-tab="${tabName}"]`);
  const activeTabPane = document.getElementById(tabName);

  if (activeTabBtn && activeTabPane) {
    activeTabBtn.classList.add('active');
    activeTabPane.classList.add('active');
  }
}

// Export Functions
function updateExportPreviews() {
  const jsonPreview = document.getElementById('jsonPreview');
  const tomlPreview = document.getElementById('tomlPreview');
  const yamlPreview = document.getElementById('yamlPreview');
  const markdownPreview = document.getElementById('markdownPreview');

  if (jsonPreview) {
    jsonPreview.textContent = JSON.stringify(classData, null, 2);
  }

  if (tomlPreview) {
    tomlPreview.textContent = convertToTOML(classData);
  }

  if (yamlPreview) {
    yamlPreview.textContent = convertToYAML(classData);
  }

  if (markdownPreview) {
    markdownPreview.textContent = convertToMarkdown(classData);
  }
}

function convertToTOML(data) {
  let toml = `[${data.name}]\n`;
  data.groups.forEach((group, index) => {
    toml += `[[${data.name}.groups]]\n`;
    toml += `name = "${group.name}"\n`;
    toml += `students = [\n`;
    group.students.forEach(student => {
      toml += `  { fullName = "${student.fullName}", userName = "${student.userName}", email = "${student.email}" },\n`;
    });
    toml += `]\n\n`;
  });
  return toml;
}

function convertToYAML(data) {
  let yaml = `${data.name}:\n`;
  yaml += `  groups:\n`;
  data.groups.forEach(group => {
    yaml += `    - name: "${group.name}"\n`;
    yaml += `      students:\n`;
    group.students.forEach(student => {
      yaml += `        - fullName: "${student.fullName}"\n`;
      yaml += `          userName: "${student.userName}"\n`;
      yaml += `          email: "${student.email}"\n`;
    });
  });
  return yaml;
}

function convertToMarkdown(data) {
  let markdown = `# ${data.name}\n\n`;
  data.groups.forEach(group => {
    markdown += `## ${group.name}\n\n`;
    if (group.students.length > 0) {
      markdown += `| Full Name | Username | Email |\n`;
      markdown += `| --- | --- | --- |\n`;
      group.students.forEach(student => {
        markdown += `| ${student.fullName} | ${student.userName} | ${student.email} |\n`;
      });
      markdown += `\n`;
    } else {
      markdown += `No students in this group.\n\n`;
    }
  });
  return markdown;
}

// Download Functions
function downloadFile(content, filename, mimeType) {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

// Event Listeners for Modal
closeModal.addEventListener('click', closeExportModal);

exportModal.addEventListener('click', (e) => {
  if (e.target === exportModal) {
    closeExportModal();
  }
});

// Tab switching
tabBtns.forEach(btn => {
  btn.addEventListener('click', () => {
    const tabName = btn.dataset.tab;
    switchTab(tabName);
  });
});

// Export button event listeners
exportJsonBtn.addEventListener('click', () => {
  const content = JSON.stringify(classData, null, 2);
  downloadFile(content, `${classData.name}.json`, 'application/json');
});

exportTomlBtn.addEventListener('click', () => {
  const content = convertToTOML(classData);
  downloadFile(content, `${classData.name}.toml`, 'application/toml');
});

exportYamlBtn.addEventListener('click', () => {
  const content = convertToYAML(classData);
  downloadFile(content, `${classData.name}.yaml`, 'application/yaml');
});

exportMarkdownBtn.addEventListener('click', () => {
  const content = convertToMarkdown(classData);
  downloadFile(content, `${classData.name}.md`, 'text/markdown');
});

exportPdfBtn.addEventListener('click', () => {
  // For PDF export, we'll use a simple approach with print styles
  // In a real application, you might use jsPDF or similar library
  const printWindow = window.open('', '_blank');
  const html = `
    <html>
      <head>
        <title>${classData.name}</title>
        <style>
          body { font-family: Arial, sans-serif; margin: 20px; }
          h1 { color: #333; }
          h2 { color: #666; margin-top: 30px; }
          table { width: 100%; border-collapse: collapse; margin-top: 10px; }
          th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
          th { background-color: #f5f5f5; }
        </style>
      </head>
      <body>
        <h1>${classData.name}</h1>
        ${classData.groups.map(group => `
          <h2>${group.name}</h2>
          ${group.students.length > 0 ? `
            <table>
              <tr><th>Full Name</th><th>Username</th><th>Email</th></tr>
              ${group.students.map(student => `
                <tr>
                  <td>${student.fullName}</td>
                  <td>${student.userName}</td>
                  <td>${student.email}</td>
                </tr>
              `).join('')}
            </table>
          ` : '<p>No students in this group.</p>'}
        `).join('')}
      </body>
    </html>
  `;

  printWindow.document.write(convertToMarkdown(classData));
  printWindow.document.close();
  printWindow.print();
});

exportImageBtn.addEventListener('click', () => {
  const selectedFormat = document.querySelector('input[name="imageFormat"]:checked').value;

  // For image export, we'll create a simple text-based representation
  // In a real application, you might use html2canvas or similar library
  const canvas = document.createElement('canvas');
  const ctx = canvas.getContext('2d');

  // Set canvas size
  canvas.width = 800;
  canvas.height = 600;

  // Fill background
  ctx.fillStyle = getComputedStyle(document.body).getPropertyValue('--color-surface');
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  // Add text
  ctx.fillStyle = getComputedStyle(document.body).getPropertyValue('--color-text');
  ctx.font = '20px Arial';
  ctx.fillText(`${classData.name} - Export`, 20, 40);

  let yPosition = 80;
  classData.groups.forEach(group => {
    ctx.font = '16px Arial';
    ctx.fillText(group.name, 20, yPosition);
    yPosition += 30;

    group.students.forEach(student => {
      ctx.font = '14px Arial';
      ctx.fillText(`${student.fullName} (${student.userName}) - ${student.email}`, 40, yPosition);
      yPosition += 25;
    });
    yPosition += 20;
  });

  // Download the canvas as image
  canvas.toBlob((blob) => {
    const url = URL.createObjectURL(blob);
    const link = document.createElement('a');
    link.href = url;
    link.download = `${classData.name}.${selectedFormat}`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    URL.revokeObjectURL(url);
  }, `image/${selectedFormat}`, 0.9);
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
