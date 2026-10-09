import { redirect } from 'next/navigation';
import { getSubAppPaths, resolveSubAppType } from './helpers/subApp';

export default function Home() {
  const subAppType = resolveSubAppType(
    process.env.NEXT_PUBLIC_SUB_APP_TYPE ?? process.env.SUB_APP_TYPE,
  );
  redirect(getSubAppPaths(subAppType).prefix);
}
