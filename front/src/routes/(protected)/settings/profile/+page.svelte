<script lang="ts">
    import { updateUser, changeEmail, changePassword } from "$lib/auth-client";
    import { useUser } from "$lib/stores/user.svelte";

    const user = useUser();

    let name = $state("");
    let email = $state("");
    let currentPassword = $state("");
    let newPassword = $state("");
    let confirmPassword = $state("");

    let profileLoading = $state(false);
    let emailLoading = $state(false);
    let passwordLoading = $state(false);

    let profileMessage = $state<{
        type: "success" | "error";
        text: string;
    } | null>(null);
    let emailMessage = $state<{
        type: "success" | "error";
        text: string;
    } | null>(null);
    let passwordMessage = $state<{
        type: "success" | "error";
        text: string;
    } | null>(null);

    // Initialize form values when session loads
    $effect(() => {
        if (user.state.user) {
            name = user.state.user.name || "";
            email = user.state.user.email || "";
        }
    });

    async function handleUpdateProfile() {
        profileMessage = null;
        profileLoading = true;

        try {
            const { error } = await updateUser({ name });
            if (error) {
                profileMessage = {
                    type: "error",
                    text: error.message || "Failed to update profile",
                };
            } else {
                profileMessage = {
                    type: "success",
                    text: "Profile updated successfully",
                };
            }
        } catch {
            profileMessage = {
                type: "error",
                text: "An unexpected error occurred",
            };
        } finally {
            profileLoading = false;
        }
    }

    async function handleUpdateEmail() {
        emailMessage = null;
        emailLoading = true;

        try {
            const { error } = await changeEmail({ newEmail: email });
            if (error) {
                emailMessage = {
                    type: "error",
                    text: error.message || "Failed to update email",
                };
            } else {
                emailMessage = {
                    type: "success",
                    text: "Verification email sent. Please check your inbox.",
                };
            }
        } catch {
            emailMessage = {
                type: "error",
                text: "An unexpected error occurred",
            };
        } finally {
            emailLoading = false;
        }
    }

    async function handleUpdatePassword() {
        passwordMessage = null;

        if (newPassword !== confirmPassword) {
            passwordMessage = { type: "error", text: "Passwords do not match" };
            return;
        }

        if (newPassword.length < 8) {
            passwordMessage = {
                type: "error",
                text: "Password must be at least 8 characters",
            };
            return;
        }

        passwordLoading = true;

        try {
            const { error } = await changePassword({
                currentPassword,
                newPassword,
            });
            if (error) {
                passwordMessage = {
                    type: "error",
                    text: error.message || "Failed to update password",
                };
            } else {
                passwordMessage = {
                    type: "success",
                    text: "Password updated successfully",
                };
                currentPassword = "";
                newPassword = "";
                confirmPassword = "";
            }
        } catch {
            passwordMessage = {
                type: "error",
                text: "An unexpected error occurred",
            };
        } finally {
            passwordLoading = false;
        }
    }
</script>

<article>
    <header>
        <h1>Profile Settings</h1>
        <p>Manage your account information</p>
    </header>

    {#if user.state.user}
        <!-- Profile Section -->
        <section>
            <header>
                <h2>Profile Information</h2>
                <p>Update your personal details</p>
            </header>

            <form
                onsubmit={(e) => {
                    e.preventDefault();
                    handleUpdateProfile();
                }}
            >
                <div class="form-group">
                    <label for="name">Name</label>
                    <input
                        type="text"
                        id="name"
                        bind:value={name}
                        placeholder="Your name"
                        autocomplete="name"
                        required
                    />
                </div>

                {#if profileMessage}
                    <div
                        class="message"
                        class:success={profileMessage.type === "success"}
                        class:error={profileMessage.type === "error"}
                    >
                        {profileMessage.text}
                    </div>
                {/if}

                <button
                    type="submit"
                    class="btn btn-primary"
                    disabled={profileLoading}
                >
                    {#if profileLoading}
                        <span class="spinner spinner-sm"></span>
                    {/if}
                    Save Changes
                </button>
            </form>
        </section>

        <!-- Email Section -->
        <section>
            <header>
                <h2>Email Address</h2>
                <p>Change your email address</p>
            </header>

            <form
                onsubmit={(e) => {
                    e.preventDefault();
                    handleUpdateEmail();
                }}
            >
                <div class="form-group">
                    <label for="email">Email</label>
                    <input
                        type="email"
                        id="email"
                        bind:value={email}
                        placeholder="your@email.com"
                        autocomplete="email"
                        required
                    />
                    <span class="form-hint"
                        >A verification email will be sent to confirm the change</span
                    >
                </div>

                {#if emailMessage}
                    <div
                        class="message"
                        class:success={emailMessage.type === "success"}
                        class:error={emailMessage.type === "error"}
                    >
                        {emailMessage.text}
                    </div>
                {/if}

                <button
                    type="submit"
                    class="btn-primary"
                    disabled={emailLoading || email === user.state.user?.email}
                >
                    {#if emailLoading}
                        <span class="spinner spinner-sm"></span>
                    {/if}
                    Update Email
                </button>
            </form>
        </section>

        <!-- Password Section -->
        <section>
            <header>
                <h2>Password</h2>
                <p>Change your password</p>
            </header>

            <form
                onsubmit={(e) => {
                    e.preventDefault();
                    handleUpdatePassword();
                }}
            >
                <!-- Hidden field for password manager autocomplete -->
                <div style="display:none;">
                    <label for="password-form-name">Name</label>
                    <input
                        type="text"
                        id="password-form-name"
                        value={name}
                        placeholder="Your name"
                        autocomplete="name"
                        readonly
                    />
                </div>
                <div class="form-group">
                    <label for="current-password">Current Password</label>
                    <input
                        type="password"
                        id="current-password"
                        bind:value={currentPassword}
                        placeholder="Enter current password"
                        autocomplete="current-password"
                        required
                    />
                </div>

                <div class="form-group">
                    <label for="new-password">New Password</label>
                    <input
                        type="password"
                        id="new-password"
                        bind:value={newPassword}
                        placeholder="Enter new password"
                        autocomplete="new-password"
                        required
                        minlength="8"
                    />
                    <span class="form-hint">Minimum 8 characters</span>
                </div>

                <div class="form-group">
                    <label for="confirm-password">Confirm New Password</label>
                    <input
                        type="password"
                        id="confirm-password"
                        bind:value={confirmPassword}
                        placeholder="Confirm new password"
                        autocomplete="new-password"
                        required
                    />
                </div>

                {#if passwordMessage}
                    <div
                        class="message"
                        class:success={passwordMessage.type === "success"}
                        class:error={passwordMessage.type === "error"}
                    >
                        {passwordMessage.text}
                    </div>
                {/if}

                <button
                    type="submit"
                    class="btn-primary"
                    disabled={passwordLoading ||
                        !currentPassword ||
                        !newPassword ||
                        !confirmPassword}
                >
                    {#if passwordLoading}
                        <span class="spinner spinner-sm"></span>
                    {/if}
                    Change Password
                </button>
            </form>
        </section>
    {:else}
        <div class="loading-state">
            <div class="spinner spinner-dark"></div>
        </div>
    {/if}
</article>
