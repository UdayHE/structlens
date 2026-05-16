import { MainLayout } from "@/layouts/main-layout";
import { AppStoreProvider } from "@/store/app-store";

function App() {
  return (
    <AppStoreProvider>
      <MainLayout />
    </AppStoreProvider>
  );
}

export default App;
