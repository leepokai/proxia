import LoginForm from "./LoginForm";

export default function LoginPage() {
  return (
    <main className="mx-auto max-w-md p-6">
      <h1 className="text-2xl font-semibold mb-2">Sign in</h1>
      <p className="text-sm text-gray-600 mb-6">Use a magic link to sign in.</p>
      <LoginForm />
    </main>
  );
}





