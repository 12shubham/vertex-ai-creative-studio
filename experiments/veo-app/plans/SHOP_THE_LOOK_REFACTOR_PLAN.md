# Refactoring Plan: `pages/shop_the_look.py`

## 1. Background and Rationale

### 1.1. Problem Analysis
The current implementation of the "Shop the Look" feature in `pages/shop_the_look.py` is a single, monolithic file exceeding 3000 lines. This file tightly couples several distinct responsibilities:
-   **Complex UI Rendering:** It contains the logic for multiple UI states, including model selection, look selection, results display, and a configuration panel.
-   **State Management:** A single, large `PageState` class manages the state for all UI components and workflow stages.
-   **Business Logic:** All logic for model interaction is embedded directly within UI event handlers (e.g., `on_click_vto_look`). This includes a complex, multi-step workflow involving numerous calls to VTO, Gemini, and Veo models, as well as concurrent operations.
-   **Data Access:** All Firestore queries and data persistence logic are co-located with the UI code.

This monolithic structure makes the feature difficult to understand, debug, maintain, and extend.

### 1.2. Goal and Architectural Inspiration
The primary goal is to refactor `pages/shop_the_look.py` into a modular, maintainable, and testable feature that aligns with the best architectural patterns observed elsewhere in this application.

Our refactoring strategy is guided by the following successful patterns:
-   **Component-Based UI (`pages/imagen.py`, `pages/veo.py`):** These pages demonstrate a clean separation of concerns, where the main page file acts as a simple container that composes smaller, single-responsibility UI components from a dedicated `components/<feature>/` directory.
-   **Isolated Model Layer (`models/image_models.py`, `models/veo.py`):** In these examples, the complex logic of interacting with generative models is abstracted into a dedicated `models/` file. The UI layer simply calls these functions, making the code easier to test and reason about.
-   **Generator-Based Workflows (`pages/character_consistency.py`):** This page provides an elegant pattern for handling long-running, multi-step processes. By using a generator (`yield`ing `WorkflowStepResult`), it provides granular, real-time status updates to the user, which is a significant UX improvement. The "Shop the Look" feature is a perfect candidate for this approach.

By adopting these patterns, we will make the "Shop the Look" feature more robust, readable, and consistent with the rest of the application.

## 2. Phased Implementation and Validation Plan

This refactoring is broken down into sequential phases. Each phase includes a checklist of tasks and a corresponding set of manual validation steps for you to perform to ensure correctness before proceeding to the next phase.

---

### **Phase 1: Foundational Refactoring (State and Data Logic)**
**Objective:** Isolate state management and data access logic from the UI code. This is the foundational step that enables further componentization.

#### Tasks:
- [x] Create a new file: `state/shop_the_look_state.py`.
- [x] Move the `PageState` class from `pages/shop_the_look.py` to the new state file.
- [x] Update `pages/shop_the_look.py` to import `PageState` from its new location.
- [x] Create a new file: `models/shop_the_look_workflow.py`.
- [x] Move the data loading and storage functions (`load_model_data`, `store_model_data`, `load_article_data`, `store_article_data`, `load_look_data`) from `pages/shop_the_look.py` to `models/shop_the_look_workflow.py`.
- [x] Update the functions in `pages/shop_the_look.py` to call the data functions from their new location in the models file.

#### Manual Validation (Phase 1):
1.  Load the "Shop the Look" page in the application.
2.  Navigate to the "Config" tab.
3.  **Verification:** Confirm that the "Apparel" and "Models" sections correctly load and display the items from your Firestore database.
4.  Navigate back to the "Shop the Look" tab.
5.  **Verification:** Confirm that the "Choose a Model" and "Choose a Look" sections correctly populate with the pre-existing models and apparel items.
6.  **Note:** At this stage, the core "Try On" functionality is not expected to work. The goal is to verify that the UI still loads its initial data correctly after the refactoring.

---

### **Phase 2: Componentizing the UI (Selection and Configuration)**
**Objective:** Decompose the main UI into smaller, more manageable components for selection and configuration, following the pattern of `imagen` and `veo`.

#### Tasks:
- [ ] Create the directory `components/shop_the_look/`.
- [ ] Create `components/shop_the_look/config_panel.py` and move the `vto_enterprise_config` component logic into it.
- [ ] Create `components/shop_the_look/model_selection.py` and move the `stl_model_select` component logic into it.
- [ ] Create `components/shop_the_look/look_selection.py` and move the `stl_look_select` component logic into it.
- [ ] Update `pages/shop_the_look.py` to import and use these new components, commenting out the old, now-redundant functions within the page file.

#### Manual Validation (Phase 2):
1.  Load the "Shop the Look" page.
2.  Navigate to the "Config" tab.
3.  **Verification:** Confirm the configuration panel renders correctly. Test changing a setting (e.g., "VTO Sample Count") and ensure the state is updated (the UI should reflect the change).
4.  Navigate to the "Shop the Look" tab.
5.  **Verification:** Confirm the "Choose a Model" screen renders correctly from its new component. Test uploading a new model image and selecting an existing model. The UI should update to the "Choose a Look" screen.
6.  **Verification:** Confirm the "Choose a Look" screen renders correctly. Test uploading new articles and selecting them. The UI should update to show the selected items.
7.  **Note:** Clicking the main "Try On" or "Continue" button is not expected to work yet.

---

### **Phase 3: Implementing the Workflow and Results UI**
**Objective:** Re-implement the core generation logic using the robust generator pattern and componentize the final results display.

#### Tasks:
- [ ] In `models/shop_the_look_workflow.py`, implement the `run_shop_the_look_workflow()` generator function, moving the complex, multi-step logic from the original `on_click_vto_look` event handler. This new function will `yield` `WorkflowStepResult` objects.
- [ ] In `models/shop_the_look_workflow.py`, create the `generate_shop_the_look_video()` function, moving the logic from the original `on_click_veo` handler.
- [ ] Create `components/shop_the_look/results_display.py` and move the `stl_result` component logic into it.
- [ ] Refactor the primary "Try On" button's click handler in `pages/shop_the_look.py`. It should now call and iterate over the new `run_shop_the_look_workflow()` generator, updating the page's status message and results from the yielded data at each step.
- [ ] Update the "Create Video" button in the new `results_display.py` component to call the new `generate_shop_the_look_video()` function.
- [ ] Update the main page to use the new `results_display.py` component.

#### Manual Validation (Phase 3):
1.  Perform a full, end-to-end test of the "Shop the Look" workflow.
2.  Select a model and a look. Click the "Try On" button.
3.  **Verification:** Observe the status text on the page. Confirm that it updates in real-time to reflect the current step of the workflow (e.g., "Trying on shirt...", "Running final critic...").
4.  **Verification:** Once the workflow completes, confirm that the final generated image and the progression image galleries are displayed correctly in the results area.
5.  **Verification:** If the final critic deems the image inaccurate, confirm the retry logic is triggered correctly.
6.  **Verification:** Click the "Create Video" button and confirm that a video is generated from the final image.
7.  **Verification:** Click the "Clear" button and confirm that the entire page state is reset correctly.

---

### **Phase 4: Final Cleanup**
**Objective:** Remove all legacy code and finalize the refactoring.

#### Tasks:
- [ ] Delete all the old, commented-out functions and classes from `pages/shop_the_look.py`.
- [ ] Review all new files in `state/`, `models/`, and `components/shop_the_look/` for code clarity, comments, and adherence to project style guides.
- [ ] Mark this plan as complete.

#### Manual Validation (Phase 4):
1.  Perform one final, end-to-end smoke test of the entire feature to ensure no regressions were introduced during the cleanup.
2.  **Verification:** The "Shop the Look" page should be fully functional and stable.
