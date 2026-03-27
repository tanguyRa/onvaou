<script lang="ts">
	import { signIn, signUp } from "$lib/auth-client";
	import { goto } from "$app/navigation";
	import { t, tError } from "$lib/i18n/index.svelte";

	let name = $state("");
	let email = $state("");
	let password = $state("");
	let error = $state("");
	let loading = $state(false);

	async function handleSubmit(e: Event) {
		e.preventDefault();
		error = "";
		loading = true;

		try {
			const result = await signUp.email({
				email,
				password,
				name,
			});

			if (result.error) {
				error = tError(
					result.error.message,
					"auth.register.errors.failed",
				);
			} else {
				const signInResult = await signIn.email({
					email,
					password,
				});

				if (signInResult.error) {
					error = tError(
						signInResult.error.message,
						"auth.login.errors.failed",
					);
					return;
				}

				goto("/app");
			}
		} catch (err) {
			error = tError(
				err instanceof Error ? err.message : null,
				"common.errors.unexpected",
			);
		} finally {
			loading = false;
		}
	}
</script>

<section>
	<header>
		<a href="/" class="auth-logo">SaaS Seed</a>
		<div>
			<h1>{t("auth.register.title")}</h1>
			<p>{t("auth.register.subtitle")}</p>
		</div>
	</header>

	{#if error}
		<div class="auth-error">{error}</div>
	{/if}

	<form onsubmit={handleSubmit}>
		<section>
			<label for="name">{t("auth.fields.name")}</label>
			<input
				type="text"
				id="name"
				bind:value={name}
				required
				placeholder={t("auth.fields.namePlaceholder")}
			/>
		</section>

		<section>
			<label for="email">{t("auth.fields.email")}</label>
			<input
				type="email"
				id="email"
				bind:value={email}
				required
				placeholder={t("auth.fields.emailPlaceholder")}
			/>
		</section>

		<section>
			<label for="password">{t("auth.fields.password")}</label>
			<input
				type="password"
				id="password"
				bind:value={password}
				required
				placeholder={t("auth.fields.passwordPlaceholder")}
				minlength="8"
			/>
		</section>

		<button type="submit" class="primary" disabled={loading}>
			{loading
				? t("auth.register.actions.creatingAccount")
				: t("auth.register.actions.createAccount")}
		</button>
	</form>

	<p class="auth-footer">
		{t("auth.register.footer.hasAccount")}
		<a href="/login">{t("auth.register.footer.signIn")}</a>
	</p>
</section>

<style>
	.auth-logo {
		font-family: var(--font-display);
		font-size: 1.5rem;
		text-align: center;
		color: var(--color-ink);
		letter-spacing: 0.02em;
	}

	.auth-error {
		background: var(--color-danger-bg);
		color: var(--color-danger);
		border: 1px solid var(--color-danger-border);
		padding: 0.75rem 1rem;
		border-radius: var(--radius-md);
		font-size: var(--font-size-sm);
	}

	.auth-footer {
		text-align: center;
		color: var(--color-text-muted);
		font-size: var(--font-size-sm);
	}

	.auth-footer a {
		color: var(--color-primary);
		font-weight: 600;
	}
</style>
