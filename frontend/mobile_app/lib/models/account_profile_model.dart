class AccountProfileModel {
  const AccountProfileModel({
    required this.accountName,
    required this.companyName,
    required this.accountEmail,
    required this.mobilePhone,
    required this.phone,
    required this.address,
    required this.companyLogoUrl,
    required this.isActive,
  });

  final String accountName;
  final String companyName;
  final String accountEmail;
  final String mobilePhone;
  final String phone;
  final String address;
  final String companyLogoUrl;
  final bool isActive;

  factory AccountProfileModel.fromJson(Map<String, dynamic> json) =>
      AccountProfileModel(
        accountName: json['account_name']?.toString() ?? '',
        companyName: json['company_name']?.toString() ?? '',
        accountEmail: json['account_email']?.toString() ?? '',
        mobilePhone: json['mobilephone']?.toString() ?? '',
        phone: json['phone']?.toString() ?? '',
        address: json['address']?.toString() ?? '',
        companyLogoUrl: json['company_logo_url']?.toString() ?? '',
        isActive: json['is_active'] == true,
      );
}
